exports.handler = function(event, context) {
  var exec = require('child_process').exec;

  var cmd = "./bam-weather"
  var child = exec(cmd, function(error, stdout, stderr) {
    if(!error) {
      console.log('stdout: ' + stdout);
      console.log('stderr: ' + stderr);
      context.done();
    } else {
      console.log('stdout: ' + stdout);
      console.log('stderr: ' + stderr);
      console.log('err code: ' + error.code + ' err: ' + error);
      context.done(error, 'lambda');
    }
  });
}
