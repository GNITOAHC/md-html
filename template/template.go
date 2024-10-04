package template

func GetTemplate() string {
	return (`
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />

    <title>Markdown Preview</title>

    <!-- Load KaTeX CSS -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/katex/dist/katex.min.css" />

    <!-- Load Highlight.js CSS -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.10.0/build/styles/default.min.css" />
    <script src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.10.0/build/highlight.min.js"></script>

    <!-- VS Code Markdown CSS -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/Microsoft/vscode/extensions/markdown-language-features/media/markdown.css" />

    <!-- Live Reload -->
    <script type="text/javascript" src="https://livejs.com/live.js"></script>
  </head>

  <body>
    {{ .Content }}

    <!-- KaTeX Render -->
    <script
      type="text/javascript"
      id="MathJax-script"
      async
      src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js"
    ></script>
    <!-- Highlight.js Render -->
    <script>hljs.highlightAll();</script>

    <style>
      body {
          font-size: 14px;
          line-height: 1.6;
      }
      .hljs {
          background: none;
      }
    </style>

  </body>
</html>
    `)
}
