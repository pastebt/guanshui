{{define "OBJ"}}
<ul><li>{{if or .Del (gt .Tid 0)}}<a target="post"
 href="/phh/post.phh?cle={{.Cle}}">--</a>{{else}}{{if .IsNew}}<img
 border=0 src="/static/new.gif" alt="new" />{{end}}{{if .Afile }}<a target="post"
href="/phh/atta.phh?cle={{.Cle}}"><img border=0 src="/static/pclip.gif"
 alt="clip" title="{{.Afile}}" /></a>{{end}}<a target="post"
 href="/phh/post.phh?cle={{.Cle}}">{{.Sub}}{{end}}</a>{{if and (not .Del) (eq .Tid 0) (gt .Size 0)}} [{{.Size}}]{{end}}
{{if .Del}}<span style="text-decoration: line-through;">{{end}}
{{if gt .Tid 0 }}<span style="color: gray;">{{end}}
-- [{{.Who}}{{if .Tow }} to {{.Tow}}{{end}}] {{.When}}
{{if gt .Visit 0}}({{.Visit}}){{end}}
{{if or .Del (gt .Tid 0)}}</span>{{else}}
{{range .Cool }}&nbsp;<img src="/static/cool.gif" />{{end}}
{{end}}
{{range .Children}}{{template "OBJ" .}}{{end}}
</li></ul>
{{end}}

{{define "NAV"}}
<center>
{{ if .Left}}<a
 href="/phh/list.phh?first={{.Left}}&new={{if .Last}}1{{else}}0{{end}}"><img
 src="/static/a-left.gif" border=0 /></a>{{end}}{{if and .Left .Right}}&nbsp;
{{end}}{{ if .Right}}<a
 href="/phh/list.phh?first={{.Right}}&new={{if .Last}}1{{else}}0{{end}}">
<img src="/static/a-right.gif" border=0 /></a>
{{end}}
</center>
{{end}}

Messages: {{.TotalNum}} &nbsp; &nbsp;
Visits: {{.VisitCnt}} &nbsp; &nbsp;
{{if .ActiveCnt}}Active: {{.ActiveCnt}}{{end}}
{{template "NAV" .}}
<div class="msg_subject">
{{range .Posts}}{{template "OBJ" .}}
{{end}}
</div>
{{template "NAV" .}}
