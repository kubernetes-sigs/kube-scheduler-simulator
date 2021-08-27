export const objectToTreeViewData = (ent: Object | null): Array<object> => {
  if (ent == null) {
    return [];
  }

  const data = [];
  for (const [key, value] of Object.entries(ent)) {
    if (typeof value == "object") {
      data.push({
        id: key,
        name: key,
        children: objectToTreeViewData(value),
      });
    } else if (Array.isArray(value)) {
      data.push({
        id: key,
        name: key,
        children: objectToTreeViewData(value),
      });
    } else {
      data.push({
        id: key,
        name: key + ": " + value,
      });
    }
  }
  return data;
};

export const schedulingResultToTableData = (result: {
  [nodeName: string]: { [name: string]: string | number };
}): Array<{ [name: string]: string | number }> => {
  const ret: Array<{ [name: string]: string | number }> = [];

  for (const nodeName in result) {
    const data: { [name: string]: string | number } = {};
    data["Node"] = nodeName;
    for (const pluginName in result[nodeName]) {
      data[pluginName] = result[nodeName][pluginName];
    }
    ret.push(data);
  }
  return ret;
};

interface header {
  text: string;
  value: string;
}

export const extractTableHeader = (result: {
  [nodeName: string]: { [name: string]: string | number };
}): Array<header> => {
  const ret: Array<header> = [
    {
      text: "Node　　　　　　　　　　　",
      value: "Node",
    },
  ];

  const added: { [name: string]: boolean } = {};
  for (const nodeName in result) {
    for (const pluginName in result[nodeName]) {
      if (!added[pluginName]) {
        ret.push({
          text: pluginName,
          value: pluginName,
        });
        added[pluginName] = true;
      }
    }
  }
  return ret;
};
