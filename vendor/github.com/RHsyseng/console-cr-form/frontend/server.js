const express = require("express");
const webpackDevMiddleware = require("webpack-dev-middleware");
const webpack = require("webpack");
const webpackConfig = require("./webpack.config.js");
const app = express();
const fs = require("fs");

const compiler = webpack(webpackConfig);

app.use(
  webpackDevMiddleware(compiler, {
    hot: true,
    filename: "dist/bundle.js",
    publicPath: "/",
    stats: {
      colors: true
    },
    historyApiFallback: true
  })
);

app.use(express.static(__dirname));

const server = app.listen(3000, function() {
  const host = server.address().address;
  const port = server.address().port;
  console.log("Test server app listening at http://%s:%s", host, port);
});

app.get('/api/form', (req, res) => {
  return res.send(readFile("full-form.json"));
});

app.get('/api/schema', (req, res) => {
  return res.send(readFile("full-schema.json"));
});

app.get('/api/spec', (req, res) => {
  return res.send({
    kind: "KieApp",
    apiVersion: "web-served.app.kiegroup.org/v1"
  });
});

app.post('/api', (req, res) => {
  let body = "";
  req.on("data", chunk => {
    body += chunk.toString();
  });
  req.on("end", () => {
    console.log("Hey I deployed this stuff: ", body);
    res.end("{\"Result\": \"Success\"}");
  });
});


function readFile(fileName) {
  return JSON.parse(fs.readFileSync("../test/examples/" + fileName));
}
