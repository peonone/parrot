var wsUri;
var output;
var count;
var ws;

function print(message) {
  $("#info").append("<li>" + new Date().toISOString() + ":" + message + "</li>");
};


function login() {
  $.post("/auth/login", $("#loginForm").serialize(), function(data) {
    if(data.success) {
      print("login success");
      $("#login").hide();
      $("#chat").show();
      connectWs(data);
    }else {
      print("login failed: " + data.errMsg);
    }
  })
}

function connectWs(data) {
    ws = new WebSocket("ws://"+window.location.host+"/chat/ws");
    ws.onopen = function(evt) {
      print('ws connected');
      ws.send(JSON.stringify({uid: data.uid, token: data.token}))
    }
    ws.onclose = function(evt) {
      print('ws closed');
      ws = null;
    }
    ws.onmessage = function(evt) {
      print('ws msg:' + evt.data);
    }
    ws.onerror = function(evt) {
      print('ws error' + evt.data);
    }
}
function getFormData($form){
  var unindexed_array = $form.serializeArray();
  var indexed_array = {};

  $.map(unindexed_array, function(n, i){
      indexed_array[n['name']] = n['value'];
  });

  return indexed_array;
}

function sendPM() {
  ws.send(JSON.stringify(getFormData($("#private"))));
}
