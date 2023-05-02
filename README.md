# SR05 Projet de programmation P23

Bordeau Oceane, Gharbi Wassim, Scheidler Nicolas, Schneider Martin

## Presentation

Notre projet permet d'accèder à un document texte en lecture/écriture depuis plusieurs sites de manière simulatanée. L'édition de ce document correspond à la section critique du programme. La communication des sites est assurée par le biais d'un anneau.
Pour répondre à nos objectifs, un algortihme réparti a été implémenté. Celui-ci est composé de plusieurs algortihmes fondamentaux :

- Un algorithme de file d'attente : Coordonner les demandes d'entrée en section critique entre les différents sites.
- Un algorithme de sauvegarde : Capturer un instantané cohérent de l'état de l'éxécution (document, demandes en attente, traffic réseau).

Une interface web est également disponible.

## Installation

Tout d'abord, il est necessaire d'importer le projet sur votre machine.

Vous pouvez cloner le repo Github :
`$ git clone https://github.com/martinschnder/projet_sr05.git`
ou décompresser l'archive tar.

Voici l'arborescence que vous devriez avoir :
![Screenshot from 2023-05-02 01-46-43.png](./_resources/Screenshot%20from%202023-05-02%2001-46-43.png)

## Lancement

Compilez le projet pour obtenir un executable :
`$ go build .`
Un éxécutable "projet" devrait apparaître.

Exécuter le script startup.sh :
`$ ./startup.sh`
Il execute une instance du projet par site et les met en réseau en anneau :
`./projet -id 0 -port 2222 < /tmp/f | ./projet -id 1 -port 3333 | ./projet -id 2 -port 4444 > /tmp/f`
Ici, nous créons 3 sites distincts sur 3 ports : 2222, 3333, 4444.

Ouvrir les interfaces web et connecter les websockets, en ouvrant autant de pages que de sites (3 ici):
`$ firefox client/client.html`

Une fois connecté, vous devriez obtenir des logs semblables :
![Screenshot from 2023-05-02 02-05-15.png](./_resources/Screenshot%20from%202023-05-02%2002-05-15.png)

## Utilisation

Avec le visuel du document, vous pouvez suivre son état. Depuis chaque site, vous pouvez éditer le document :

- Ligne : Selectionner le numéro de la ligne que vous souhaitez modifier
- Action : Choisisser ce que vous voulez faire sur cette ligne (Remplace, Ajouter ou Supprimer)
- Texte : Entrer le texte que vous voulez

Vous pouvez vérifier qu'un action sur un site entraîne la modification de chaque réplicats sur tous les sites.

Enfin vous pouvez réaliser un snapshot. L'appui du bouton généra un fichier "snapshot.txt".

## Implémentation des algorithmes

### File d'attente

Le code a été écrit en se basant sur l'algorithme de file d'attente fourni (lien).

La file d'attente se base sur les deux attributs de la classe net :

- `clock int ` : l'horloge logique du site
- `tab [NB_SITES]Request` : Etat des requetes de demandes d'accès en section critique

A partir de l'algorithme du cours, des méthodes ont été écrites :

- `func (n *Net) ReceiveCSrequest()` : Envoi d'une requete d'accès au document partagé
- `func (n *Net) receiveCSRelease()` : Envoi d'une requete de libération du document partagé
- `func (n *Net) receiveRequestMessage(received_msg Message)` : Traitement d'une requete d'accès au document partagé reçue
- `func (n *Net) receiveReleaseMessage(msg Message)` : Traitement d'un message de liberation du document partagé reçu
- `func (n *Net) receiveAckMessage(msg Message)` : traitement d'un message d'acquittement reçu.
- `func (n *Net) isLastRequest() bool` : Comparative des estampilles pour valider que le site a emis la quete la plus vieille et donc peut accéder à la section critique
- `func (n *Net) writeMessage(msg Message)` : Diffusion d'un message dans l'anneau

Enfin, nous avons pu implémenter le programme en déroulant l'algorithme du cours.

## Snapshot

Le code a été écrit en se basant sur l'algorithme d'instantané avec reconstitution de configuration (algo 11 du cours 6).

Les snapshots se basent sur les attributs suivants de la classe net :

- `color string` : Etat du snapshot
- `initator bool` : Site demandant un snapshot (faux si c'est un autre site)
- `nbExpectedStates int` : Nombre de sites du réseau
- `nbExpectedMessages int`: Nombre de messages envoyés mais non reçu
- `state *State`:
- `globalStates *list.List`:

A partir de l'algorithme 11, des méthodes ont été écrites :

- `func (n *Net) receiveSnapshotMessage(msg Message)` : Demande de création d'un instantané.
- `func (n *Net) InitSnapshot()` : Début de l'instantané
- `func (n *Net) receiveSnapshotMessage(msg Message)` :
- `func (n *Net) receiveStateMessage(msg Message)` : Reception d'un message de type State
- `func (n *Net) receivePrepostMessage(msg Message)` : Réception d’un message prépost
- `func (n *Net) reinitializeAfterSnapshot()` :
- `func (n *Net) receiveEndSnapshotMessage()`:
- `func (n *Net) SendMessageFromServer(msg Message)` :

Enfin, nous avons pu implémenter le code en déroulant l'algorithme 11.

## Documentation

### Package utils : Utilitaires

- Fonctions des logs
  - `func Info(id int, where string, what string)` : Log d'informations
  - `func Warning(id int, where string, what string)` : Log d'avertissement
  - `func Error(id int, where string, what string)` : Log d'erreur
- Variable de
  - `var rouge string`
  - `var orange string`
  - `var raz string`
  - `var cyan string`
  - `var pid``
  - `var stderr`

### Package net : Class Net et Server

- net.go
  - Attributs
  ```
  	type Net struct {
  id       int
  clock    int
  tab      [NB_SITES]Request
  messages chan MessageWrapper
  server   *Server
  color 	 string
  initator bool
  nbExpectedStates int
  nbExpectedMessages int
  state 	*State
  globalStates *list.List
  }
  ```
  - Methodes
    - `NewNet(id int, port string, addr string) *Net` : Constructeur
    - `func (n *Net) ReceiveCSrequest()` : Envoi d'une requete d'accès au document partagé
    - `func (n *Net) receiveCSRelease()` : Envoi d'une requete de libération du document partagé
    - `func (n *Net) receiveExternalMessage(msg Message)` : Appel de la fonction approprié selon le type du message
    - `func (n *Net) receiveRequestMessage(received_msg Message)` : Traitement d'une requete d'accès au document partagé reçue
    - `func (n *Net) isLastRequest() bool` : Comparative des estampilles pour valider que le site a emis la quete la plus vieille et donc peut accéder à la section critique
    - `func (n *Net) receiveReleaseMessage(msg Message)` : Traitement d'un message de liberation du document partagé reçu
    - `func (n *Net) receiveAckMessage(msg Message)` : traitement d'un message d'acquittement reçu.
    - `func (n *Net) receiveSnapshotMessage(msg Message)` :
    - `func (n *Net) receiveStateMessage(msg Message)` :
    - `func (n *Net) receivePrepostMessage(msg Message)` :
    - `func (n *Net) reinitializeAfterSnapshot()` :
    - `func (n *Net) receiveEndSnapshotMessage()`:
    - `func (n *Net) ReadMessage()` :
    - `func (n *Net) writeMessage(msg Message)` :
    - `func (n *Net) MessageHandler()` :
    - `func (n *Net) SendMessageFromServer(msg Message)` :
    - `func (n *Net) InitSnapshot()` :
- server.go
  - Attributs
  ```
  	type Server struct {
  socket *websocket.Conn
  mutex *sync.Mutex
  data []string
  id int
  net *Net
  command Command
  }
  ```
  - Methodes
    - `func NewServer(port string, addr string, id int, net *Net) *Server : Constructeur
    - `func (server *Server)createSocket(w http.ResponseWriter, r *http.Request) : Creation de la socket
    - `func (server *Server)Send()` :
    - `func (server *Server)closeSocket()` :
    - `func (server *Server)receive()` :
    - `func (server *Server)EditData(command Command)` :
    - `func (server *Server)forwardEdition(command Command)` :
    - `func (server *Server)SendMessage(action string)` :

### Package Message : Class State

- Type :

  ```
  type Message struct {
  	From        int
  	To          int
  	Content     string
  	Stamp       int
  	MessageType string
  	VectClock 	[]int
  	Color 		string
  }
  ```

  ```
  type MessageWrapper struct {
  	Action  string
  	Message Message
  }
  ```

  ```
  type Request struct {
  	RequestType string
  	Stamp       int
  }
  ```

  ```
  type Command struct {
    Line int
    Action string
    Content string
  }
  ```

  ```
  type State struct {
  Id int
  VectClock []int
  Text []string
  Review int
  }
  ```

- Fonctions :
  - `func MessageFromString(raw string) Message` : Formatte une string sous forme de message pour les entrées standards.
  - `func vectClockToString(vectClock []int)` string : Transforme une horloge en string
  - `func vectClockToArray(raw string) []int` : Horloge sous forme de tableau
  - `func (msg Message) ToString() string` : Message sous forme de string
  - `func (msg Message) ConcernSnapshot() bool` :
  - `func (msg Message) Send()` : Envoi d'un message
  - `func ParseCommand(raw string) Command` : Parse une string pour la mettre sous forme de command
  - `func (command Command) ToString() string` : fonction réciproque de ParseCommand(raw string)
  - `func NewState(id int, text []string, nbSites int) *State` : Constructeur
  - `func (s *State)VectClockIncr(otherClock []int, nbSites int)` :
  - `func (s *State) ToString() string` :
  - `func textToString(text []string) string` :
  - `func textFromString(raw string) []string` :
  - `func StateFromString(raw string) State` :
