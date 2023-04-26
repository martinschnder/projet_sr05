mkfifo /tmp/in_F
./projet -time 2 < /tmp/F | ./projet -time 3 > /tmp/F

# Fonction de nettoyage
nettoyer () {
  # Suppression des processus de l'application app
  killall projet 2> /dev/null

  # Suppression des processus de l'application ctl
  killall ctl 2> /dev/null
 
  # Suppression des processus tee et cat
  killall tee 2> /dev/null
  killall cat 2> /dev/null
 
  # Suppression des tubes nommés
  \rm -f /tmp/in* /tmp/out*
  exit 0
}
 
# Appel de la fonction nettoyer à la réception d'un signal
echo "+ INT QUIT TERM => nettoyer"
trap nettoyer INT QUIT TERM
