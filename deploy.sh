#!/bin/bash
multitail \
    -z \
    -cT ansi -l 'git push ancientcitadelgifs1 master' \
    -cT ansi -l 'git push ancientcitadelgifs2 master' \
    -cT ansi -l 'git push ancientcitadelgifs3 master' \
    -cT ansi -l 'git push ancientcitadelgifs4 master' \
    -cT ansi -l 'git push ancientcitadelgifs5 master'
