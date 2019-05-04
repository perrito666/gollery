package render

/*
MIT License

Copyright (c) 2019 Horacio Duran <horacio.duran@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

var singleImageDefaultTemplate = []byte(`<html>

<head>
    <title>{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</title>
    <meta name="description" content="{{.Description}}">

    <meta property="og:title" content="{{if .Title}}{{.Title}}{{else}}{{.FileName}}{{end}}" />
    <meta property="og:type" content="website" />
    <meta property="og:description" content="{{if .Description}}{{.Description}}{{end}}" />
    {{if .Metadata.baseURL}}
    {{/* if you want this, create a json file with a dict containing baseURL and pass it to --extrameta */}}
    <meta property="og:url" content="{{.Metadata.baseURL}}{{.RelativePath}}" />
    <meta property="og:image" content="{{.Metadata.baseURL}}{{.ThumbName 320 213}}" />
    {{end}}
</head>
</body>
<nav class="breadcrumb">
    <p>
    {{if .ParentTree}}
        {{range .ParentTree}}
        <a href="{{.TraversePath}}">{{if .Title}} {{.Title}} {{else}} {{.FolderName}} {{end}}</a>/
        {{end}}
    {{end}}
    </p>
</nav>
<nav class="tree">
    {{/* Device a logic for current so we can add a class */}}
    {{range .Siblings}}
    <a href="{{.TraversePath}}" class="sibling_album {{if .Current}} active{{end}}">{{if .Title}} {{.Title}} {{else}} {{.FolderName}} {{end}}</a>
    {{end}}
    {{range .Children}}
    <a href="{{.TraversePath}}" class="children_album">{{if .Title}} {{.Title}} {{else}} {{.FolderName}} {{end}}</a>
    {{end}}
</nav>
<nav class="relative">
    {{if .Previous}}
    <a href="{{.First.FileName}}" class="first_picture">First</a>
    {{end}}
    {{if .Previous}}
    <a href="{{.Previous.FileName}}" class="previous_picture">Previous</a>
    {{end}}
    {{if .Next}}
    <a href="{{.Next.FileName}}" class="next_picture">Next</a>
    {{end}}
    {{if .Next}}
    <a href="{{.Last.FileName}}" class="last_picture">Last</a>
    {{end}}
</nav>
<header>
    <h1>{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</h1>
</header>
<main>
    <img src="{{.FileName}}" alt="{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}">
    <p>{{.Description}}</p>
</main>
<footer></footer>
</body>
</html>
`)

var groupDefaultTemplate = []byte(`<html>

<head>
    <title>{{if .Title}} {{.Title}} {{else}} {{.FolderName}} {{end}}</title>
    <meta name="description" content="{{.Description}}">

    <meta property="og:title" content="{{if .Title}}{{.Title}}{{else}}{{.FolderName}}{{end}}" />
    <meta property="og:type" content="website" />
    <meta property="og:description" content="{{if .Description}}{{.Description}}{{end}}" />
    {{if .Metadata.baseURL}}
    <meta property="og:url" content="{{.Metadata.baseURL}}{{.RelativePath}}" />
    {{end}}
</head>
<body>
<nav class="breadcrumb">
    <p>
    {{if .ParentTree}}
    {{range $i, $e := .ParentTree}}
    <a class="breadcrumb-item {{if LastItem $i (len $.ParentTree)}}active{{end}}"
        {{if LastItem $i (len $.ParentTree)}}aria-current="page" {{end}}
        href="{{$e.TraversePath}}">{{if $e.Title}} {{$e.Title}} {{else}} {{$e.FolderName}} {{end}}</a>
    {{end}}
    {{end}}
    </p>
</nav>
<nav class="tree">
    {{range .Siblings}}
    <a href="{{.FolderName}}" class="sibling_album {{if .Current}} active{{end}}">{{if .Title}} {{.Title}} {{else}} {{.FolderName}} {{end}}</a>
    {{end}}
    {{range .Children}}
    <a href="{{.FolderName}}" class="sibling_album">{{if .Title}} {{.Title}} {{else}} {{.FolderName}} {{end}}</a>
    {{end}}
</nav>
<header>
    <h1>{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</h1>
</header>
<main>
    {{range .Images}}
    <section class="thumb">
        <a href="{{.FileName}}">
            {{/* thumbs are obtaining by GETing FileName___widthxheight assuming said size is permitted if it is not you just get the default one */}}
            <img src="{{.ThumbName 320 213}}" alt="{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}">
            <p>{{.Description}}{{/* perhaps this could be truncated */}}</p>
        </a>
    </section>
    {{end}}
</main>
<footer></footer>
</body>
</html>`)
