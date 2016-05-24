var EventEmitter = require('wolfy87-eventemitter');


exports.GoSock = function (url, options) {
  this.url = url;
  this.ee = new EventEmitter();
  this.ws = new WebSocket(url);
  this.connected = this.ws.readyState === 1 ? true : false;
  this.options = {
    retries: 5,
    retryInterval: 3000,
  };

  if (options) {
    this.options = Object.assign(this.options, options);
  }

  this.on = function(event, callback) {
    this.ee.addListener(event, callback);
  }

  this._emit = function(event, data) {
    this.ee.emitEvent(event, [data]);
  }

  this.ws.onopen(function(){
    this._emit('open', null);
  });

  this.ws.onmessage(function(data){
    var delimIndex = data.indexOf(' ')
    var event = data.substr(0, delimIndex);

    try {
      data = JSON.parse(data.substr(delimIndex + 1));
      this._emit(data.substr(0, delimIndex), data.substr(delimIndex + 1));
    } catch() {
      this._emit(data.substr(0, delimIndex), data.substr(delimIndex + 1));
    }
  });

  this.ws.onclose = function() {
    this._emit('close', null);
    this.connected = 0;
    this._reconnect();
  }

  this._reconnect = function() {
    if (this.options.retries > 0) {
      counter = 0;
      var retry = setInterval(function(){
        if (counter === this.options.retries) {
          clearInterval(retry);
        }

        if (!this.connected) {
          this.ws = new WebSocket(this.url);
          if (this.ws.readyState === 1) {
            this.connected = true;
            clearInterval(retry);
          }
        }

        counter++;
      }, this.options.retryInterval);
    }
  }

  this.ws.onerror = function(err) {
    this._emit('error', err);
  }

  this.close = function() {
    this.ws.close();
    this.connected = 0;
  }

  this.remove = function(event) {
    this.ee.removeEvent(event);
  }

  this.send(event, data) {
    switch (typeof data) {
      case 'object':
        this.ws.send(event + ' ' + JSON.stringify(data));
        break;
      default:
        this.ws.send(event + ' ' + data);
    }
  }
};
