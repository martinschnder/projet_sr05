var ws;
var fieldsep = "/";
var keyvalsep = "=";

document.getElementById("connecter").onclick = function (evt) {
  if (ws) {
    return false;
  }

  var host = document.getElementById("host").value;
  var port = document.getElementById("port").value;

  addToLog("Attempting to connect to server");
  addToLog("host = " + host + ", port = " + port);
  ws = new WebSocket("ws://" + host + ":" + port + "/ws");

  ws.onopen = function (evt) {
    addToLog("Connection established");
  };

  ws.onclose = function (evt) {
    addToLog("Connection closed");
    ws = null;
  };

  ws.onmessage = function (evt) {
    addToLog("Receiving data from server");
    let jsonMessage = JSON.parse(evt.data);
    const editor = document.querySelector(".editor");
    editor.innerHTML = "";
    jsonMessage.Data.forEach(function (line, index) {
      const lineDiv = document.createElement("div");
      lineDiv.classList.add("line");

      const lineNumber = document.createElement("span");
      lineNumber.textContent = index + 1;
      lineNumber.classList.add("line-number");

      lineDiv.appendChild(lineNumber);

      const lineContent = document.createTextNode(line);
      lineDiv.appendChild(lineContent);

      editor.appendChild(lineDiv);
    });
    document.getElementById("hlg").innerHTML = jsonMessage.Stamp;
  };

  ws.onerror = function (evt) {
    addToLog("Erreur: " + evt.data);
  };
  return false;
};

document.getElementById("fermer").onclick = function (evt) {
  if (!ws) {
    return false;
  }
  ws.close();
  return false;
};

document.getElementById("envoyer").onclick = function (evt) {
  if (!ws) {
    return false;
  }
  var line = document.getElementById("select-line-number").value;
  var action = document.getElementById("select-action").value;
  var message = document.getElementById("text").value;
  var sndmsg =
    format("line", line) +
    format("action", action) +
    format("message", message);

  addToLog("Sending command to server");
  ws.send(sndmsg);
  return false;
};

document.getElementById("snapshot").onclick = function (evt) {
  if (!ws) {
    return false;
  }

  var sndmsg =
    format("line", "") +
    format("action", "Snapshot") +
    format("message", "");

  addToLog("Requesting a snapshot to server");
  ws.send(sndmsg);
  return false;
};

const selectAction = document.getElementById("select-action");
const textInput = document.getElementById("text");

selectAction.addEventListener("change", function () {
  if (selectAction.value === "Supprimer") {
    textInput.disabled = true;
    textInput.value = "";
  } else {
    textInput.disabled = false;
  }
});

function addToLog(message) {
  var logs = document.getElementById("text-logs");
  var d = document.createElement("div");
  d.textContent = message;
  logs.appendChild(d);
  logs.scroll(0, logs.scrollHeight);
}

function format(key, val) {
  return fieldsep + keyvalsep + key + keyvalsep + val;
}

