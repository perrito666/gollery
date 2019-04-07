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
    <description>{{.Description}}</description>
</head>
<nav class="breadcrumb">
    <p>
        {{range .ParentTree}}
        <a href="{{.RelativePath}}">{{.FolderName}}</a>/
        {{end}}
    </p>
</nav>
<nav class="tree">
    {{/* Device a logic for current so we can add a class */}}
    {{range .Siblings}}
    <a href="{{.FileName}}" class="sibling_picture{{if .Current}} current_image{{end}}">{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</a>
    {{end}}
    {{range .Children}}
    <a href="{{.FolderName}}" class="sibling_album">{{if .Title}} {{.Title}} {{else}} {{.FolderName}} {{end}}</a>
    {{end}}
</nav>
<nav class="relative">
    <a href="{{.First}}" class="first_picture">{{if .First.Title}} {{.First.Title}} {{else}} {{.First.FileName}}
        {{end}}</a>
    <a href="{{.Previous}}" class="previous_picture">{{if .Previous.Title}} {{.Previous.Title}} {{else}}
        {{.Previous.FileName}} {{end}}</a>
    <a href="{{.Next}}" class="next_picture">{{if .Next.Title}} {{.Next.Title}} {{else}} {{.Next.FileName}} {{end}}</a>
    <a href="{{.Last}}" class="last_picture">{{if .Last.Title}} {{.Last.Title}} {{else}} {{.Last.FileName}} {{end}}</a>
</nav>
<header>
    <h1>{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</h1>
</header>
<main>
    <img src="{{.FileName}}" alt="{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}">
    <p>{{.Description}}</p>
</main>
<footer></footer>

</html>
`)

var groupDefaultTemplate = []byte(`<html>

<head>
    <title>{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</title>
    <description>{{.Description}}</description>
</head>
<nav class="breadcrumb">
    <p>
        {{range .ParentTree}}
        <a href="{{.RelativePath}}">{{.FolderName}}</a>/
        {{end}}
    </p>
</nav>
<nav class="tree">
    {{/* Device a logic for current so we can add a class */}}
    {{range .Siblings}}
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
            <img src="{{.FileName}}___640x480" alt="{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}">
            <p>{{.Description}}{{/* perhaps this could be truncated */}}</p>
        </a>
    </section>
    {{end}}
</main>
<footer></footer>

</html>`)
