export function getJsonSchemaPathForJsonPath(jsonPath) {
  //console.log("json Path: " + jsonPath);
  jsonPath = jsonPath.slice(2, jsonPath.length);
  jsonPath = jsonPath.replace(/\./g, ".properties.");
  jsonPath = "$.." + jsonPath;

  //console.log("jsonSchema Path: " + jsonPath);
  return jsonPath;
}

export function getJsonSchemaPathForYaml(jsonPath) {
  //console.log("json Path: " + jsonPath);
  jsonPath = jsonPath.slice(2, jsonPath.length);

  //console.log("jsonSchema Path: " + jsonPath);
  return jsonPath;
}

export function replaceStarwithPos(field, jsonPath) {
  //  console.log("field: " + field);
  //if(field.jsonpath.search(/env\[\*\]/g)!=-1)
  var envPos = document.getElementById(field.id).getAttribute("envpos");
  if (envPos != undefined) {
    var str1 = "env[" + envPos + "]";
    // var res = str.replace(/env\[\*\]/g, str1);
    jsonPath = jsonPath.replace(/env\[\*\]/g, str1);
  }
  //jsonPath = jsonPath.slice(2, jsonPath.length);

  //  console.log("jsonSchema Path:*********************** " + jsonPath);
  return jsonPath;
}
