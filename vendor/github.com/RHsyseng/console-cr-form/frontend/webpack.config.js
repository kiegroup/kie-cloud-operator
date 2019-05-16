const path = require("path");

module.exports = {
  mode: "development",
  context: path.join(__dirname, "src"),
  entry: ["./index.js"],
  output: {
    path: path.join(__dirname, "dist"),
    filename: "bundle.js",
    publicPath: '/'
  },
  devServer: {
    contentBase: path.join(__dirname, "dist"),
    hot: true,
    proxy: {
      "/api": {
        target: "http://localhost:3000",
        secure: false
      }
    }
  },
  module: {
    rules: [
      {
        test: /\.(png|jpg|gif)$/i,
        use: [
          {
            loader: "url-loader",
            options: {
              limit: 8192
            }
          }
        ]
      },
      {
        test: /\.(js|jsx)$/,
        exclude: /(node_modules|__test__)/,
        use: ["babel-loader"]
      },
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: ["babel-loader", "eslint-loader"]
      },
      {
        test: /\.css$/,
        use: ["style-loader", "css-loader"]
      },
      {
        test: /\.(svg|ttf|eot|woff|woff2)$/,
        use: {
          loader: "file-loader",
          options: {
            // Limit at 50k. larger files emited into separate files
            limit: 5000,
            outputPath: "fonts",
            name: "[name].[ext]"
          }
        }
      }
    ]
  },
  node: {
    fs: "empty"
  }
};
