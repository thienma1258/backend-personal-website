#git submodule update --init --recursive
git submodule foreach git reset --hard HEAD #reset
git submodule foreach git pull origin master
