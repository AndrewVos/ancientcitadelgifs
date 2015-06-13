#!/bin/bash

multitail \
    -cT ansi -l 'heroku logs --tail -a ancientcitadelgifs1' \
    -cT ansi -l 'heroku logs --tail -a ancientcitadelgifs2' \
    -cT ansi -l 'heroku logs --tail -a ancientcitadelgifs3' \
    -cT ansi -l 'heroku logs --tail -a ancientcitadelgifs4' \
    -cT ansi -l 'heroku logs --tail -a ancientcitadelgifs5'
