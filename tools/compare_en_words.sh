#!/usr/bin/env bash

US_DICT="../game/words/en_us"

# For example on install british dict
# apt-get install wbritish-large

SYSTEM_GB_DICT='/usr/share/dict/british-english-large'
OUTPUT_GB_DICT="../game/words/en_gb"
FIXLIST="./fixlist"

rm -f ${FIXLIST} ${OUTPUT_GB_DICT}

while IFS= read -r line
do 
  grep -x ${line} ${SYSTEM_GB_DICT} # Check it exists in GB list
  if [[ $? == 0 ]];
    then
      echo ${line} >> ${OUTPUT_GB_DICT} # Write to en_gb
    else
      echo ${line} >> ${FIXLIST} # Finally write to fixlist if no matches
  fi
done < "${US_DICT}"
