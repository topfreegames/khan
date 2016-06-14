export PMD=pmd-bin-5.4.2
export PMDUrl="http://downloads.sourceforge.net/project/pmd/pmd/5.4.2/pmd-bin-5.4.2.zip?r=https\%3A\%2F\%2Fsourceforge.net\%2Fprojects\%2Fpmd\%2Ffiles\%2Fpmd\%2F5.4.2\%2F&ts=1465934375&use_mirror=tenet"

cd "/tmp"

if [ ! -d $PMD ]; then
  curl -o $PMD.zip -L -O $PMDUrl
  [ -e $PMD.zip ] && unzip $PMD.zip
fi
