package consts

const (
	MessageCardTemplateWithTitle1 = `
{
  "config": {
    "autoLayout": true,
    "enableForward": true
  },
  "contents": [
    {
      "type": "markdown",
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

	MessageCardTemplateWithoutTitle = `
{
  "config": {
    "autoLayout": true,
    "enableForward": true
  },
   "contents": [
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

	ReceivedTypeText  = "text"
	ReceivedTypeImage = "picture"
	ReceivedTypeVoice = "audio"
)

var VoicePrefix = []string{}
