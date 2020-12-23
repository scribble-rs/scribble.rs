#!/usr/bin/env bash

if [ -z ${FROM+x} ]; then echo "FROM is unset"; exit 1; fi
if [ -z ${TO+x} ]; then echo "TO is unset"; exit 1; fi

SOURCE_FILE="../resources/words/${FROM}"
DESTINATION_FILE="../resources/words/${TO}"

if [ ! -f ${SOURCE_FILE} ]; then echo "file: ${SOURCE_FILE} does not exist"; fi;
if [ -f ${DESTINATION_FILE} ]; then
    echo "WARNING! file: ${DESTINATION_FILE} already exist!";
    echo "Do you want to override it? [yes/no]"
    read -r choice
    if [[ "${choice}" != "yes" ]]; then
        echo "Aborting"
        exit 1;
    fi

    rm -f ${DESTINATION_FILE}
fi;

touch ${DESTINATION_FILE}

while read sourceWord; do
  cleanWord=$(echo "$sourceWord" | sed -En "s/(.*)\#.*/\1/p")
  tag=$(echo "$sourceWord" | sed -En "s/.*\#(.*)/\1/p")

  echo -n "Translating '${cleanWord}'... "

  # Wanna exclude some words based on a tag?
  # Just un-comment and edit the following lines
  if [[ ${tag} == "i" ]]; then
    echo "❌ Skipping due to tag setting."
    continue;
  fi

  # non-optimized AWS call
  # Must use a translation-job here
  translation=$(aws translate translate-text --text "${cleanWord}" --source-language-code "${FROM}" --target-language-code "${TO}" | jq -r .TranslatedText)

  echo "${translation}" >> ${DESTINATION_FILE}
  echo "✅"
done <${SOURCE_FILE}

echo "Done!"
