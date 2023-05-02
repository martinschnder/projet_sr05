## Explication des premières avancées :

- Déroulement des étapes sur la page [moodle](https://moodle.utc.fr/mod/page/view.php?id=155982) jusqu'à l'étape _recaler l'horloge locale à l'arrivée_.

---

**Point sur startup.sh : j'arrive pas à le lancer **

**Dans le main :**
Création des sites (avec package flag) : un site est une instance de la classe "net".
Dans le code, le site local correspond à la variable "n".
Pour cette notice, le "siteA" correspondra toujours au site local.
Le "siteB" sera l'expéditeur des messages recus.
Le "siteC" sera le destinataire des messages recus.

Lancement goroutines : app.ReadMessage() et app.MessageHandler()

**Goroutine app.ReadMessage():** Pour initier un message
Dans une boucle :

1.  Attente une saisie (un message du siteA) en boucle sur l'entrée standard (stdin) : la string "raw"
2.  "raw" est mis sous notre format de message (type Message) par la fonction MessageFromString
3.  Après verification que le destinateur n'est pas l'expediteur (un message précedemment envoyé serait revenu après parcours dans l'anneau), envoi dans le chan (bizarre enft le 3.)

**Goroutine app.MessageHandler():** Pour la réception et le traitement selon la situation.
Dans une boucle :

1.  Lecture bloquante dans le chan du siteA, en attente d'un message arrivant. Le traitement du message est fait par le biais de la variable wrapperItem (type MessageWrapper)
2.  Selon la valeur de la valeur de l'attribut "Action" de wrapperItem:
    - Si Action = "send": (le message va tourner dans l'anneau jusqu'à arriver au destinataire)
      Diffusion du message dans l'anneau par la fonction message.Send() : ecriture du message sur la sortie standart (stdout)
    - Si Action = "process" : Traitement par net.receiveExternalMessage()
      - le message est pour le siteA (msg.To == n.id) :
        on traite uniquement le message
      - le message est pour tout le monde (msg.To == -1 && msg.From != n.id) :
        on traite le message, et on le diffuse sur l'anneau avec message.Send()
      - le message n'est pas pour le siteA :
        on le diffuse uniquement avec message.Send()
3.  Traitement du message, selon le type du message
    - LockRequestMessage : fct receiveRequestMessage()
      A. On fait jsp trop quoi avec n.clock()
      B. On actualise le statut de la Request en "access" du siteB l'acces exclusive dans l'attribut tab du siteA. Le siteA considere alors que la ressource est utilisé par le siteB et qu'elle lui ait indisponible.
      C. On envoie une confirmation au siteB avec un message de type "AckMessage" et l'action "Send" (garanti de reception par le siteB) avec net.writeMessage() (construit le MessageWrapper approprié)
      D. **_Si tata alors TODO send okCS to server_**
    - ReleaseMessage : fct receiveReleaseMessage()
      A. On fait jsp trop quoi avec n.clock()
      B. On actualise le statut de la Request en "release" du siteB demander dans l'attribut tab du siteA. Le site considere alors que la ressource n'est pas utilisé par le siteB.
      C. **_Si tata alors TODO send okCS to server_**
    - AckMessage : fct receiveAckMessage() :
      A. On fait jsp trop quoi avec n.clock()
      B. Si le site considérait que l'expéditeur n'utilisait pas la ressource, il actualise le statut de l'expiditeur en "ack". Le siteA considère que le siteB lui accorde la ressource.
      C. **_Si tata alors TODO send okCS to server_**

**Type:**

```
type Net struct {
    id       int
    clock    int
    tab      [NB_SITES]Request //implemente la connaissance du monde du site id.
    messages chan MessageWrapper
    server   Server
}
```

- Message : structure tel que

```
type Message struct {
    From        int
    To          int
    Content     string
    Stamp       int
    MessageType string
}
```

- MessageWrapper :

```
type MessageWrapper struct {
    Action  string ("send" || "process")
    Message Message
}
```

- MessageType :
  - Lock Request Message (Message de demande de verrouillage )
  - Ack Message (Message de confirmation)
  - Release Message (Message de libération)
- Request :

```
type Request struct {
    RequestType string
    Stamp       int
}
```

- Server :

```
type Server struct {
  socket *websocket.Conn
  mutex *sync.Mutex
  data []string
  id int
  net *Net
}
```

Questions :
Comment fonctionne MessageFromString() ?
Quel est l'interet de n.logger() ?
Qu'indique l'attribut "stamp" ? Lien entre stamp et clock mais quoi exactement ?
Quel est le role du serveur exactement ?
