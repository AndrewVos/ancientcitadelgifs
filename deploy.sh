#!/bin/bash

MACHINES=5
printf $(seq -s , 1 $MACHINES) |
    xargs -d ',' -n 1 -P 4 -I{} git push ancientcitadelgifs{} master
