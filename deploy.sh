COMMANDS='
git push heroku master
git push ancientcitadelgifs1 master
git push ancientcitadelgifs2 master
git push ancientcitadelgifs3 master
git push ancientcitadelgifs4 master
git push ancientcitadelgifs5 master
'
echo "$COMMANDS" | tr '\n' '\0' | xargs --null -n 1 -P 6 bash -c
