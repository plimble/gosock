GoSock client
-----

## Usage
```
var sock = new GoSock('ws://localhost', {
  reconnect: true,
  retries: 5,
  retryInterval: 1000,
});

sock.on('open', function(){
  console.log('open');
});

sock.on('error', function(err){
  console.log(err);
});

sock.on('close', function(){
  console.log('close');
});

sock.on('new.msg.response', function(data){
  console.log(data);
});

sock.send('new.msg', {data: 'hello world'});
```
