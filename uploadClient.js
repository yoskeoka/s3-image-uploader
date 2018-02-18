// 呼び出し方
// node uploadClient.js api-url idToken

if (process.argv.length < 2) {
  console.error("IDトークンを指定して下さい");
  process.exit(1);
}

const apiURL = process.argv[2]
const idToken = process.argv[3];
const headers = { Authorization: idToken };

const axios = require("axios");
const fs = require("fs");
const base64 = require("base64-arraybuffer");

const buf = fs.readFileSync("ninja.png");
const str = base64.encode(buf.buffer);
console.log(str);

axios
  .post(
    apiURL,
    {
      mime_type: "image/png",
      content: str
    },
    { headers }
  )
  .then(res => {
    console.log(res);
  })
  .catch(ex => {
    console.log(ex);
  });
