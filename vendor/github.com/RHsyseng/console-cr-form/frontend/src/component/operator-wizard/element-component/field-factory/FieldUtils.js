export class FieldUtils {
  static generateIds(pageNumber, fieldNumber, label) {
    const randomNum = Math.floor(Math.random() * 100000000 + 1);
    const fieldGroupId =
      pageNumber + "-fieldGroup-" + fieldNumber + "-" + label + "-" + randomNum;
    const fieldId =
      pageNumber + "-field-" + fieldNumber + "-" + label + "-" + randomNum;
    return {
      fieldGroupId: fieldGroupId,
      fieldGroupKey: "fieldGroupKey-" + fieldGroupId,
      fieldId: fieldId,
      fieldKey: "fieldKey-" + fieldId
    };
  }
}
