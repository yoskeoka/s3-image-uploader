// 呼び出し方
// node uploadClient.js idToken

if (process.argv.length < 2) {
    console.error("IDトークンを指定して下さい");
    process.exit(1);
}

const idToken = process.argv[2]
const headers = { Authorization: idToken };

const axios = require("axios");
const fs = require("fs");
const base64 = require("base64-arraybuffer");

const buf = fs.readFileSync("ninja.png");
const str = base64.encode(buf.buffer);
console.log(str);

axios.post("https://t7c56w1aqf.execute-api.us-east-1.amazonaws.com/dev/upload", {
    "mime_type": "image/png",
    "content": str
}, { headers }).then((res)=>{
    console.log(res);
}).then((err)=>{
    console.log(err);
}).catch(ex=>{
    console.log(ex);
});
