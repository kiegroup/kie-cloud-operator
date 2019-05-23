export class FieldUtils {
  static generateIds(pageNumber, fieldNumber, label, parentId, grandParentId) {
    const randomNum = Math.floor(Math.random() * 100000000 + 1);
    let fieldGroupId =
      pageNumber + "-fieldGroup-" + fieldNumber + "-" + label + "-" + randomNum;
    let fieldId =
      pageNumber + "-field-" + fieldNumber + "-" + label + "-" + randomNum;
    if (parentId !== undefined) {
      fieldGroupId =
        pageNumber +
        "_fieldGroup_" +
        fieldNumber +
        "_" +
        parentId +
        "_" +
        grandParentId +
        "_" +
        label +
        "_" +
        randomNum;

      fieldId =
        pageNumber +
        "_field_" +
        fieldNumber +
        "_" +
        parentId +
        "_" +
        grandParentId +
        "_" +
        label +
        "_" +
        randomNum;
    }
    return {
      fieldGroupId: fieldGroupId,
      fieldGroupKey: "fieldGroupKey-" + fieldGroupId,
      fieldId: fieldId,
      fieldKey: "fieldKey-" + fieldId
    };
  }
}
