package consts

const (
	MessageCardTemplate = `
{
  "config": {
    "autoLayout": true,
    "enableForward": true
  },
  "header": {
    "title": {
      "type": "text",
      "text": "流输出模式"
    },
    "logo": "@lALPDfJ6V_FPDmvNAfTNAfQ"
  },
  "contents": [
    {
      "type": "text",
      "text": "%s",
      "id": "text_1693929551595"
    },
    {
      "type": "divider",
      "id": "divider_1693929551595"
    },
    {
      "type": "markdown",
      "text": "%s",
      "id": "markdown_1693929674245"
    }
  ]
}
`
)

const (
	OutputTypeText     = "Text"
	OutputTypeStream   = "Stream"
	OutputTypeMarkDown = "MarkDown"
)
