package report

var reportHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>{{.Scenario}} Run {{.RunID}}</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4"></script>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:system-ui,sans-serif;background:#0d1117;color:#c9d1d9;padding:2rem}
h1{color:#58a6ff;font-size:1.8rem}
.sub{color:#8b949e;margin:.3rem 0 1rem}
.badge{display:inline-block;padding:.5rem 1.5rem;border-radius:8px;font-size:2rem;font-weight:bold;margin:1rem 0}
.good{background:#238636;color:#fff}
.warn{background:#d29922;color:#000}
.bad{background:#da3633;color:#fff}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:1rem;margin:1.5rem 0}
.card{background:#161b22;border:1px solid #30363d;border-radius:8px;padding:1.2rem}
.card h3{font-size:.8rem;color:#8b949e;text-transform:uppercase;margin-bottom:.4rem}
.card .v{font-size:1.8rem;font-weight:600}
.chart{background:#161b22;border:1px solid #30363d;border-radius:8px;padding:1.2rem;margin:1.5rem 0}
table{width:100%;border-collapse:collapse;margin:1rem 0}
th,td{text-align:left;padding:.6rem;border-bottom:1px solid #30363d}
th{color:#8b949e;font-size:.85rem}
</style>
</head>
<body>
<h1>{{.Scenario}}</h1>
<p class="sub">Run {{.RunID}} | {{.StartedAt.Format "2006-01-02 15:04:05"}} | Duration: {{.Duration}}</p>
<div class="badge {{if ge .Score 80}}good{{else if ge .Score 50}}warn{{else}}bad{{end}}">SCORE: {{.Score}}/100</div>
<p class="sub">{{.ScoreLine}}</p>
<div class="grid">
<div class="card"><h3>Requests</h3><div class="v">{{.Requests}}</div></div>
<div class="card"><h3>Successes</h3><div class="v" style="color:#3fb950">{{.Successes}}</div></div>
<div class="card"><h3>Failures</h3><div class="v" style="color:#da3633">{{.Failures}}</div></div>
<div class="card"><h3>p95 Latency</h3><div class="v">{{printf "%.0f" .Latencies.P95}}ms</div></div>
<div class="card"><h3>p99 Latency</h3><div class="v">{{printf "%.0f" .Latencies.P99}}ms</div></div>
<div class="card"><h3>Avg Latency</h3><div class="v">{{printf "%.0f" .Latencies.Avg}}ms</div></div>
</div>
{{if .DBStats}}
<div class="grid">
<div class="card"><h3>DB In Use</h3><div class="v">{{.DBStats.InUse}}</div></div>
<div class="card"><h3>DB Wait Count</h3><div class="v">{{.DBStats.WaitCount}}</div></div>
<div class="card"><h3>DB Wait Duration</h3><div class="v">{{.DBStats.WaitDuration}}</div></div>
</div>
{{end}}
{{if .BatchStats}}
<div class="grid">
<div class="card"><h3>Fast p95</h3><div class="v">{{printf "%.0f" .BatchStats.FastP95}}ms</div></div>
<div class="card"><h3>Slow p95</h3><div class="v">{{printf "%.0f" .BatchStats.SlowP95}}ms</div></div>
<div class="card"><h3>Batch Progress</h3><div class="v">{{.BatchStats.Done}}/{{.BatchStats.Total}}</div></div>
</div>
{{end}}
{{if .HPAStats}}
<div class="grid">
<div class="card"><h3>Current Replicas</h3><div class="v">{{.HPAStats.CurrentReplicas}}</div></div>
<div class="card"><h3>Desired Replicas</h3><div class="v">{{.HPAStats.DesiredReplicas}}</div></div>
</div>
{{end}}
<div class="chart"><h3 style="color:#8b949e;margin-bottom:1rem">RPS and Latency Over Time</h3><canvas id="tsChart" height="100"></canvas></div>
<div class="chart"><h3 style="color:#8b949e;margin-bottom:1rem">Status Code Distribution</h3><canvas id="scChart" height="60"></canvas></div>
<h3 style="margin:2rem 0 1rem">Status Codes</h3>
<table><tr><th>Status</th><th>Count</th></tr>{{range $code, $count := .StatusDist}}<tr><td>{{$code}}</td><td>{{$count}}</td></tr>{{end}}</table>
<script>
var ts={{.Timeseries}};
if(ts&&ts.length>0){new Chart(document.getElementById('tsChart'),{type:'line',data:{labels:ts.map(function(d){return d.elapsed_s.toFixed(0)+'s'}),datasets:[{label:'RPS',data:ts.map(function(d){return d.rps}),borderColor:'#58a6ff',yAxisID:'y',tension:.3,pointRadius:2},{label:'p95 ms',data:ts.map(function(d){return d.latency_p95_ms}),borderColor:'#f0883e',yAxisID:'y1',tension:.3,pointRadius:2}]},options:{responsive:true,scales:{y:{position:'left',ticks:{color:'#8b949e'},grid:{color:'#21262d'}},y1:{position:'right',ticks:{color:'#8b949e'},grid:{drawOnChartArea:false}},x:{ticks:{color:'#8b949e'},grid:{color:'#21262d'}}},plugins:{legend:{labels:{color:'#c9d1d9'}}}}})}
var sd={{.StatusDist}};
if(sd){var lbl=Object.keys(sd),vals=Object.values(sd),cols=lbl.map(function(l){var c=parseInt(l);return c>=200&&c<300?'#3fb950':c>=400?'#d29922':'#da3633'});new Chart(document.getElementById('scChart'),{type:'bar',data:{labels:lbl.map(function(l){return l==='0'?'Err':l}),datasets:[{data:vals,backgroundColor:cols}]},options:{responsive:true,plugins:{legend:{display:false}},scales:{y:{ticks:{color:'#8b949e'},grid:{color:'#21262d'}},x:{ticks:{color:'#8b949e'},grid:{color:'#21262d'}}}}})}
</script>
</body>
</html>`

var indexHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Workshop Reports</title>
<style>
body{font-family:system-ui,sans-serif;background:#0d1117;color:#c9d1d9;padding:2rem}
h1{color:#58a6ff;margin-bottom:1.5rem}
table{width:100%;border-collapse:collapse}
th,td{text-align:left;padding:.8rem;border-bottom:1px solid #30363d}
th{color:#8b949e;font-size:.85rem;text-transform:uppercase}
a{color:#58a6ff;text-decoration:none}
a:hover{text-decoration:underline}
.g{color:#3fb950}.w{color:#d29922}.b{color:#da3633}
</style>
</head>
<body>
<h1>Workshop Reports</h1>
<table>
<tr><th>Scenario</th><th>Run</th><th>Time</th><th>Score</th><th>Report</th></tr>
{{range .}}<tr>
<td>{{.Scenario}}</td>
<td>{{.RunID}}</td>
<td>{{.Time}}</td>
<td class="{{if ge .Score 80}}g{{else if ge .Score 50}}w{{else}}b{{end}}">{{.Score}}/100</td>
<td><a href="{{.Link}}">View</a></td>
</tr>{{end}}
</table>
</body>
</html>`
