# Client exchange protocol

Always contains two root keys: `structure` and `data`.

### `structure`
Contains structural elements, React nodes.

### `data`
Contains related data and styling.

```js
{
  structure: {
    type:"Document",
    documentType:"document",
    componentProperties: {},
    themeData: {},
    children: [
      {
        componentType: "wysiwyg.viewer.components.ClipArt",
        dataQuery: "#c21nx", // This is where the data, directly contained by this component comes from
        id: "iaggachh", // Layout ID
        propertyQuery: "cqgi", // Get Additional Properties from Properties data
        skin: "wysiwyg.viewer.skins.photo.NoSkinPhoto", //?
        styleId: "ca1", // Styling
        type: "Component", // Layout type
        layout: { //css adjustments
          fixedPosition: <bool>,
          height: <double>,
          width: <double>,
          x: <double>,
          y: <double>
        }
      }
    ]
  },
  data: {
    document_data: {},
    theme_data: {},
    component_data: {}
  }
}
```
