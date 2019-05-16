export function getJsonSchemaPathForJsonPath(jsonPath) {
  jsonPath = jsonPath.slice(2, jsonPath.length);
  jsonPath = jsonPath.replace(/\./g, ".properties.");
  jsonPath = "$.." + jsonPath;

  return jsonPath;
}

export function getJsonSchemaPathForYaml(jsonPath) {
  return jsonPath.slice(2, jsonPath.length);
}

export function replaceStarwithPos(field, jsonPath) {
  var envPos = document.getElementById(field.id).getAttribute("envpos");
  if (envPos != undefined) {
    var str1 = "env[" + envPos + "]";
    jsonPath = jsonPath.replace(/env\[\*\]/g, str1);
  }
  return jsonPath;
}
