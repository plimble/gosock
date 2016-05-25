var EventEmitter = require('wolfy87-eventemitter');


exports.GoSock = function (url, options) {
  var self = this;
  self.url = url;
  self.ee = new EventEmitter();
  self.ws = new WebSocket(url);
  self.connected = self.ws.readyState === 1 ? true : false;
  self.options = {
    retries: 5,
    retryInterval: 3000,
  };

  if (options) {
    self.options = Object.assign(self.options, options);
  }

  self.on = function(event, callback) {
    self.ee.addListener(event, callback);
  }

  self._emit = function(event, data) {
    self.ee.emitEvent(event, [data]);
  }

  self.ws.onopen = function(){
    self._emit('open', null);
  };

  self.ws.onmessage = function(result){
    var data = result.data;
    var delimIndex = data.indexOf(' ')
    var event = data.substr(0, delimIndex);

    try {
      data = JSON.parse(data.substr(delimIndex + 1));
      self._emit(event, data);
    } catch(err) {
      self._emit(event, data.substr(delimIndex + 1));
    }
  };

  self.ws.onclose = function() {
    self._emit('close', null);
    self.connected = 0;
    self._reconnect();
  }

  self._reconnect = function() {
    if (self.options.retries > 0) {
      counter = 0;
      var retry = setInterval(function(){
        if (counter === self.options.retries) {
          clearInterval(retry);
        }

        if (!self.connected) {
          self.ws = new WebSocket(self.url);
          if (self.ws.readyState === 1) {
            self.connected = true;
            clearInterval(retry);
          }
        }

        counter++;
      }, self.options.retryInterval);
    }
  }

  self.ws.onerror = function(err) {
    self._emit('error', err);
  }

  self.close = function() {
    self.ws.close();
    self.connected = 0;
  }

  self.remove = function(event) {
    self.ee.removeEvent(event);
  }

  self.send = function(event, data) {
    switch (typeof data) {
      case 'object':
        self.ws.send(event + ' ' + JSON.stringify(data));
        break;
      default:
        self.ws.send(event + ' ' + data);
    }
  }
};
