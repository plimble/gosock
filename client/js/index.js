
exports.GoSock = function (url) {
  this.url = url;
  this.events = {};
  this.ws = new WebSocket(url);

  this.on = function(event, callback) {
    if (this.events[event]) {
      this.events[event].push(callback);
    } else {
      this.events[event] = [callback];
    }
  }

  this.emit = function(event, data) {
    if (this.events[event]) {
      for (var i = 0; i < this.events[event].length; i++) {
        this.events[event][i](data);
      }
    }
  }

  this.ws.onopen(function(){
    this.emit('open', null);
  });

  this.ws.onmessage(function(data){
    var delimIndex = data.indexOf(' ')
    var event = data.substr(0, delimIndex);
    this.emit(data.substr(0, delimIndex), data.substr(delimIndex + 1));
  });

  this.ws.onclose = function() {
    this.emit('close', null);
  }

  this.ws.onerror = function(err) {
    this.emit('error', err);
  }

  this.close = function() {
    this.ws.close();
  }
};
