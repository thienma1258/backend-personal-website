#!/usr/bin/env groovy

import java.text.SimpleDateFormat
import java.util.*
import groovy.json.*


def JENKINS_CONFIG


// Declarative Pipeline
// https://jenkins.io/doc/book/pipeline/syntax/#declarative-pipeline
pipeline {

  agent any

  options {
    timeout(time: 2, unit: 'HOURS')
    disableConcurrentBuilds()
  }

  // https://github.com/jenkinsci/pipeline-model-definition-plugin/wiki/Parametrized-pipelines
  // parameters {
  // }

  stages {

    stage('Setup') { steps { wrap([$class: 'AnsiColorBuildWrapper', 'colorMapName': 'xterm']) { script {

      wrap([$class: 'BuildUser']) { script {
        try {
          env.BUILD_USER = "${BUILD_USER}"
          env.BUILD_USER_ID = "${BUILD_USER_ID}"
          env.BUILD_USER_EMAIL = "${BUILD_USER_EMAIL}"
        } catch (error) {
          env.BUILD_USER = 'Jenkins'
          env.BUILD_USER_ID = 'jenkins'
          env.BUILD_USER_EMAIL = 'itunes.dev@notabasement.com'
        }
      }}

      env.GIT_AUTHOR = sh label: 'Find Git Author',
        returnStdout: true,
        script: 'git --no-pager show --format="%aN <%aE>" | head -n 1'

      def JENKINS_CONFIG_JSON_STRING = readFile(file:"${WORKSPACE}/jenkins.config.json")
      JENKINS_CONFIG = new JsonSlurperClassic().parseText(JENKINS_CONFIG_JSON_STRING)

      sh "printenv | sort"

    }}}}

    stage('Integration') { steps { wrap([$class: 'AnsiColorBuildWrapper', 'colorMapName': 'xterm']) { script {

      sh "./pipeline/clean"
      sh "./pipeline/install"
      sh "./pipeline/lint"

      boolean TESTED = false

      JENKINS_CONFIG.testEnvkey.each { BRANCH_PATTERN, TEST_ENVKEY_CREDENTIALS ->

        if (BRANCH_NAME ==~ /$BRANCH_PATTERN/) {

          echo "Matched '${BRANCH_PATTERN}'"

          JENKINS_CONFIG.testEnvkey[BRANCH_PATTERN].each { TEST_ENVKEY_CREDENTIAL ->

            if (!TEST_ENVKEY_CREDENTIAL) {
              echo "No TEST_ENVKEY credential found."
              return
            }

            withCredentials([string(credentialsId: TEST_ENVKEY_CREDENTIAL, variable: 'ENVKEY')]) {
              sh "./pipeline/test"
              TESTED = true
            }

          }
        }
      }

      if (!TESTED) {
        sh "./pipeline/test"
      }

    }}}}

    stage('Delivery') { steps { wrap([$class: 'AnsiColorBuildWrapper', 'colorMapName': 'xterm']) { script {

      boolean DELIVERED = false

      JENKINS_CONFIG.deployEnvkey.each { BRANCH_PATTERN, DEPLOY_ENVKEY_CREDENTIALS ->

        if (BRANCH_NAME ==~ /$BRANCH_PATTERN/) {

          echo "Matched '${BRANCH_PATTERN}'"

          JENKINS_CONFIG.deployEnvkey[BRANCH_PATTERN].each { DEPLOY_ENVKEY_CREDENTIAL ->

            if (!DEPLOY_ENVKEY_CREDENTIAL) {
              echo "No DEPLOY_ENVKEY credential found."
              return
            }

            withCredentials([string(credentialsId: DEPLOY_ENVKEY_CREDENTIAL, variable: 'DEPLOY_ENVKEY')]) {
              sh "./pipeline/deliver"
              DELIVERED = true
            }

          }
        }
      }

      if (!DELIVERED) {
        sh "./pipeline/build"
      }

    }}}}

    stage('Deployment') { steps { wrap([$class: 'AnsiColorBuildWrapper', 'colorMapName': 'xterm']) { script {

      JENKINS_CONFIG.deployEnvkey.each { BRANCH_PATTERN, DEPLOY_ENVKEY_CREDENTIALS ->

        if (BRANCH_NAME ==~ /$BRANCH_PATTERN/) {

          echo "Matched '${BRANCH_PATTERN}'"

          JENKINS_CONFIG.deployEnvkey[BRANCH_PATTERN].each { DEPLOY_ENVKEY_CREDENTIAL ->

            if (!DEPLOY_ENVKEY_CREDENTIAL) {
              echo "No DEPLOY_ENVKEY credential found."
              return
            }

            withCredentials([string(credentialsId: DEPLOY_ENVKEY_CREDENTIAL, variable: 'DEPLOY_ENVKEY')]) {
              sh "./pipeline/deploy"
            }

          }
        }
      }

    }}}}
  }

  post {

    failure { wrap([$class: 'AnsiColorBuildWrapper', 'colorMapName': 'xterm']) { script {

      load("./lib/jenkins/slack.groovy").sendFailure(
        channel: JENKINS_CONFIG.slack.channel
      )

    }}}

    success { wrap([$class: 'AnsiColorBuildWrapper', 'colorMapName': 'xterm']) { script {

      load("./lib/jenkins/slack.groovy").sendSuccess(
        channel: JENKINS_CONFIG.slack.channel
      )

    }}}
  }
}
