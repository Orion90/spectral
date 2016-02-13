var info = require('aiff-file-info')

info.infoByFilename('./test.aiff',function(err,info){
  if (err) throw err;
  console.log(info)
})
