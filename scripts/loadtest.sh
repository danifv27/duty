#!/usr/local/bin/bash

# (id, url, duration, rate)
# declare -a TARGET1=("GET http://localhost:4567/actuator/set?name=var&id=200" "1" "1s")
declare -a TARGET2=("GET http://localhost:4567/v1/200" "2" "10s")
# declare -a TARGET3=("GET http://localhost:4567/actuator/set?name=var&id=500" "1" "1s")
declare -a TARGET4=("GET http://localhost:4567/v1/500" "2" "5s") 
# declare -a TARGET5=("GET http://localhost:4567/actuator/set?name=var&id=400" "1" "1s") 
declare -a TARGET6=("GET http://localhost:4567/v1/401" "3" "2s") 

# declare -a TARGETS=("TARGET1" "TARGET2" "TARGET3" "TARGET4" "TARGET5" "TARGET6")
declare -a TARGETS=("TARGET2" "TARGET4" "TARGET6")
#### OPTIONS REPORT 📝
TYPE="hist[0,100ms,200ms,300ms]"

for target in "${TARGETS[@]}"; do
    declare -n lst="$target"
    URL=${lst[0]}
    RATE=${lst[1]}
    DURATION=${lst[2]}
    echo "#### TARGET: $target URL: ${URL},  RATE: ${RATE}, DURATION: ${DURATION}"
    mkdir -p loadtest/metrics/$target loadtest/plots/$target loadtest/results/$target
    cd loadtest
    echo $URL | vegeta attack -rate=$RATE -duration=$DURATION | tee results.bin | vegeta report
    mv results.bin results-$target-$RATE-$DURATION.bin
    vegeta report -type=json results-$target-$RATE-$DURATION.bin > metrics-$target-$RATE-$DURATION.json
    cat results-$target-$RATE-$DURATION.bin | vegeta plot > plot-$target-$RATE-$DURATION.html
    cat results-$target-$RATE-$DURATION.bin | vegeta report -type=$TYPE
    echo "#### Complete" && echo ""

    #### Organize
    echo "" && echo "#### Organizing.. "
	mv -f *$target*.json metrics/*$target*
	mv -f *$target*.html plots/*$target*
	mv -f *$target*.bin results/*$target*
    cd ..
    echo "#### Complete Loadtest" && echo "" && echo ""
done