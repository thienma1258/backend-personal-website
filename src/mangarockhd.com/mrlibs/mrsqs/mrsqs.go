package mrsqs

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const AWS_REGION_US_WEST_1 = "us-east-1"

var sqsLock sync.RWMutex
var sqsClient *sqs.SQS

var cacheQueueUrl = make(map[string]string)
var cacheQueueUrlLock sync.RWMutex

func Initialize(awsAccessKeyID, awsSecretAccessKey string) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(AWS_REGION_US_WEST_1),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""),
	})
	if err != nil {
		log.Panic("error creating AWS SQS session: ", err)
	}
	sqsLock.Lock()
	sqsClient = sqs.New(sess)
	sqsLock.Unlock()
}

// GetQueueUrl get sqs queue URL from queue name
func GetQueueUrl(queueName string) string {
	cacheQueueUrlLock.RLock()
	queueUrl := cacheQueueUrl[queueName]
	cacheQueueUrlLock.RUnlock()
	if len(queueUrl) > 0 {
		return queueUrl
	}
	sqsLock.Lock()
	result, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	sqsLock.Unlock()
	if err != nil {
		log.Panic("GetQueueUrl error when get queue URL %v ", err)
		return ""
	}

	cacheQueueUrlLock.Lock()
	cacheQueueUrl[queueName] = *result.QueueUrl
	cacheQueueUrlLock.Unlock()
	return *result.QueueUrl
}

func GetSqsClient() *sqs.SQS {
	return sqsClient
}

func ReceiveMessage(
	queueName string,
	maxNumberOfMessages int64,
	messageAttributeName string,
	waitTimeSeconds int64,
) (*sqs.ReceiveMessageOutput, error) {
	if len(queueName) == 0 {
		return nil, errors.New("Queue name not found")
	}

	if maxNumberOfMessages < 0 || maxNumberOfMessages > 10 {
		return nil, errors.New("MaxNumberOfMessage valid values: 1 to 10")
	}

	result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(GetQueueUrl(queueName)),
		MaxNumberOfMessages:   aws.Int64(maxNumberOfMessages),
		MessageAttributeNames: aws.StringSlice([]string{messageAttributeName}),
		WaitTimeSeconds:       aws.Int64(waitTimeSeconds),
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func SendBinaryJobToQueue(queueName string, payload []byte) (*sqs.SendMessageOutput, error) {
	// get queue url
	queueURL := GetQueueUrl(queueName)

	result, err := sqsClient.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"payload": &sqs.MessageAttributeValue{
				DataType:    aws.String("Binary"),
				BinaryValue: payload,
			},
		},
		MessageBody: aws.String("{}"),
		QueueUrl:    &queueURL,
	})
	if err != nil {
		log.Printf("SendMessageToQueue error while send message %v", err)
		return &sqs.SendMessageOutput{}, err
	}
	return result, nil
}

func SendBatchBinaryJobToQueue(queueName string, payload []([]byte)) ([]*sqs.SendMessageBatchOutput, error) {
	// get queue url
	queueURL := GetQueueUrl(queueName)

	message := new(sqs.SendMessageBatchInput)
	message.QueueUrl = &queueURL

	total := len(payload)
	var results []*sqs.SendMessageBatchOutput
	var entries []*sqs.SendMessageBatchRequestEntry
	payloadLength := 0
	for i := 0; i < total; i++ {
		entry := new(sqs.SendMessageBatchRequestEntry)
		entry.Id = aws.String(queueName + "-" + strconv.Itoa(int(time.Now().UnixNano())))
		entry.MessageAttributes = map[string]*sqs.MessageAttributeValue{
			"payload": &sqs.MessageAttributeValue{
				DataType:    aws.String("Binary"),
				BinaryValue: payload[i],
			}}
		entry.MessageBody = aws.String("{}")
		entries = append(entries, entry)
		payloadLength += len(payload[i])

		if len(entries) >= 10 || payloadLength > 200000 || i == total-1 {
			log.Printf("number of message batch %d", i)
			message.Entries = entries
			result, err := sqsClient.SendMessageBatch(message)
			if err != nil {
				log.Printf("SendBatchBinaryJobToQueue error while send message %v", err)
				return nil, err
			}
			entries = []*sqs.SendMessageBatchRequestEntry{}
			payloadLength = 0
			results = append(results, result)
		}
	}

	return results, nil
}

func SendJobToQueue(queueName string, payload string) (*sqs.SendMessageOutput, error) {
	// get queue url
	queueURL := GetQueueUrl(queueName)

	result, err := sqsClient.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(payload)),
		QueueUrl:    &queueURL,
	})
	if err != nil {
		log.Printf("SendMessageToQueue error while send message %v", err)
		return &sqs.SendMessageOutput{}, err
	}
	return result, nil
}

func DeleteMessage(queueName string, receiptHandle *string) error {
	// get queue url
	queueURL := GetQueueUrl(queueName)

	_, err := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		log.Printf("Error while delete message %v", err)
		return err
	}
	return nil
}

func DeleteMessageBatch(queueName string, receiptHandles []*string) error {
	// get queue url
	queueURL := GetQueueUrl(queueName)

	messages := new(sqs.DeleteMessageBatchInput)
	messages.QueueUrl = &queueURL
	total := len(receiptHandles)
	var entries []*sqs.DeleteMessageBatchRequestEntry
	for i := 0; i < total; i++ {
		entry := new(sqs.DeleteMessageBatchRequestEntry)
		entry.Id = aws.String(queueName + "-" + strconv.Itoa(int(time.Now().UnixNano())))
		entry.ReceiptHandle = receiptHandles[i]
		entries = append(entries, entry)
		if len(entries) >= 10 || i == total-1 {
			messages.Entries = entries
			_, err := sqsClient.DeleteMessageBatch(messages)
			if err != nil {
				log.Printf("SendBatchBinaryJobToQueue error while send message %v", err)
				return err
			}
			entries = []*sqs.DeleteMessageBatchRequestEntry{}
		}
	}
	return nil
}
