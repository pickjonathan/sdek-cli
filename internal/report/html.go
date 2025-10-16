package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
)

// GenerateHTML generates an interactive HTML report from a JSON report file
func GenerateHTML(jsonPath, outputPath string) error {
	// Read the JSON report
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON report: %w", err)
	}

	// Parse the report
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("failed to parse JSON report: %w", err)
	}

	// Generate HTML
	html := generateHTMLContent(report)

	// Write to file
	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

func generateHTMLContent(report Report) string {
	// Convert report to JSON for embedding
	reportJSON, _ := json.Marshal(report)

	tmpl := template.Must(template.New("report").Parse(htmlTemplate))
	var buf strings.Builder

	data := struct {
		ReportJSON template.JS
		Title      string
	}{
		ReportJSON: template.JS(reportJSON), // Use template.JS for safe JavaScript embedding
		Title:      "SDEK Compliance Report",
	}

	tmpl.Execute(&buf, data)
	return buf.String()
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            overflow: hidden;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }

        .header p {
            opacity: 0.9;
            font-size: 1.1em;
        }

        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            padding: 30px;
            background: #f8f9fa;
        }

        .summary-card {
            background: white;
            padding: 25px;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
            transition: transform 0.2s;
        }

        .summary-card:hover {
            transform: translateY(-4px);
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
        }

        .summary-card h3 {
            color: #666;
            font-size: 0.9em;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 10px;
        }

        .summary-card .value {
            font-size: 2.5em;
            font-weight: bold;
            color: #667eea;
        }

        .summary-card .label {
            color: #999;
            font-size: 0.9em;
            margin-top: 5px;
        }

        .tabs {
            display: flex;
            background: white;
            border-bottom: 2px solid #e0e0e0;
            padding: 0 30px;
        }

        .tab {
            padding: 15px 30px;
            cursor: pointer;
            border-bottom: 3px solid transparent;
            transition: all 0.3s;
            font-weight: 500;
            color: #666;
        }

        .tab:hover {
            color: #667eea;
            background: #f8f9fa;
        }

        .tab.active {
            color: #667eea;
            border-bottom-color: #667eea;
        }

        .content {
            padding: 30px;
        }

        .framework {
            margin-bottom: 30px;
            border: 1px solid #e0e0e0;
            border-radius: 12px;
            overflow: hidden;
        }

        .framework-header {
            background: #f8f9fa;
            padding: 20px;
            cursor: pointer;
            display: flex;
            justify-content: space-between;
            align-items: center;
            transition: background 0.3s;
        }

        .framework-header:hover {
            background: #e9ecef;
        }

        .framework-title {
            font-size: 1.5em;
            font-weight: 600;
            color: #333;
        }

        .compliance-badge {
            padding: 8px 16px;
            border-radius: 20px;
            font-weight: 600;
            font-size: 0.9em;
        }

        .compliance-high {
            background: #d4edda;
            color: #155724;
        }

        .compliance-medium {
            background: #fff3cd;
            color: #856404;
        }

        .compliance-low {
            background: #f8d7da;
            color: #721c24;
        }

        .framework-body {
            display: none;
            padding: 20px;
        }

        .framework.expanded .framework-body {
            display: block;
        }

        .controls-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }

        .control-card {
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            padding: 15px;
            cursor: pointer;
            transition: all 0.3s;
        }

        .control-card:hover {
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
            transform: translateY(-2px);
        }

        .control-id {
            font-weight: 600;
            color: #667eea;
            margin-bottom: 8px;
        }

        .control-title {
            font-size: 0.95em;
            color: #333;
            margin-bottom: 10px;
        }

        .control-stats {
            display: flex;
            gap: 15px;
            font-size: 0.85em;
            color: #666;
        }

        .risk-indicator {
            display: inline-block;
            width: 12px;
            height: 12px;
            border-radius: 50%;
            margin-right: 5px;
        }

        .risk-green { background: #28a745; }
        .risk-yellow { background: #ffc107; }
        .risk-red { background: #dc3545; }

        .evidence-section {
            margin-top: 20px;
        }

        .evidence-item {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 10px;
            border-left: 4px solid #667eea;
        }

        .evidence-item.ai-enhanced {
            border-left-color: #28a745;
            background: #f0f9f4;
        }

        .evidence-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }

        .ai-badge {
            background: #28a745;
            color: white;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 0.75em;
            font-weight: 600;
        }

        .confidence-bar {
            height: 6px;
            background: #e0e0e0;
            border-radius: 3px;
            overflow: hidden;
            margin: 10px 0;
        }

        .confidence-fill {
            height: 100%;
            background: linear-gradient(90deg, #667eea, #764ba2);
            transition: width 0.5s;
        }

        .filters {
            display: flex;
            gap: 15px;
            margin-bottom: 20px;
            flex-wrap: wrap;
        }

        .filter-btn {
            padding: 8px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 20px;
            background: white;
            cursor: pointer;
            transition: all 0.3s;
            font-size: 0.9em;
        }

        .filter-btn:hover {
            border-color: #667eea;
            color: #667eea;
        }

        .filter-btn.active {
            background: #667eea;
            color: white;
            border-color: #667eea;
        }

        .search-box {
            padding: 12px 20px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 1em;
            width: 100%;
            max-width: 400px;
            margin-bottom: 20px;
        }

        .search-box:focus {
            outline: none;
            border-color: #667eea;
        }

        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.7);
            z-index: 1000;
            padding: 20px;
            overflow-y: auto;
        }

        .modal.active {
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .modal-content {
            background: white;
            border-radius: 16px;
            max-width: 900px;
            width: 100%;
            max-height: 90vh;
            overflow-y: auto;
            padding: 30px;
        }

        .modal-close {
            float: right;
            font-size: 1.5em;
            cursor: pointer;
            color: #999;
        }

        .modal-close:hover {
            color: #333;
        }

        @media (max-width: 768px) {
            .summary {
                grid-template-columns: 1fr;
            }

            .controls-grid {
                grid-template-columns: 1fr;
            }

            .header h1 {
                font-size: 1.8em;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🛡️ SDEK Compliance Dashboard</h1>
            <p>Compliance Monitoring & Evidence Analysis</p>
        </div>

        <div class="summary" id="summary"></div>

        <div class="tabs">
            <div class="tab active" onclick="switchTab('overview')">Overview</div>
            <div class="tab" onclick="switchTab('frameworks')">Frameworks</div>
            <div class="tab" onclick="switchTab('findings')">Findings</div>
            <div class="tab" onclick="switchTab('evidence')">Evidence</div>
        </div>

        <div class="content">
            <input type="text" class="search-box" id="searchBox" placeholder="🔍 Search controls, findings, or evidence..." onkeyup="handleSearch()">
            
            <div id="overview-tab" class="tab-content"></div>
            <div id="frameworks-tab" class="tab-content" style="display:none;"></div>
            <div id="findings-tab" class="tab-content" style="display:none;"></div>
            <div id="evidence-tab" class="tab-content" style="display:none;"></div>
        </div>
    </div>

    <div class="modal" id="detailModal">
        <div class="modal-content">
            <span class="modal-close" onclick="closeModal()">&times;</span>
            <div id="modalBody"></div>
        </div>
    </div>

    <script>
        const reportData = {{.ReportJSON}};
        let currentFilter = 'all';

        function init() {
            renderSummary();
            renderOverview();
            renderFrameworks();
            renderFindings();
            renderEvidence();
        }

        function renderSummary() {
            const summary = document.getElementById('summary');
            const totalControls = reportData.summary.total_controls;
            const totalEvidence = reportData.summary.total_evidence;
            const totalFindings = reportData.summary.total_findings;
            const aiAnalyzed = countAIEvidence();
            
            summary.innerHTML = ` + "`" + `
                <div class="summary-card">
                    <h3>Frameworks</h3>
                    <div class="value">${reportData.frameworks.length}</div>
                    <div class="label">Monitored</div>
                </div>
                <div class="summary-card">
                    <h3>Controls</h3>
                    <div class="value">${totalControls}</div>
                    <div class="label">Total Controls</div>
                </div>
                <div class="summary-card">
                    <h3>Evidence</h3>
                    <div class="value">${totalEvidence}</div>
                    <div class="label">${aiAnalyzed} AI-Enhanced</div>
                </div>
                <div class="summary-card">
                    <h3>Findings</h3>
                    <div class="value">${totalFindings}</div>
                    <div class="label">Issues Found</div>
                </div>
            ` + "`" + `;
        }

        function renderOverview() {
            const container = document.getElementById('overview-tab');
            let html = '<h2 style="margin-bottom: 20px;">Compliance Overview</h2>';
            
            reportData.frameworks.forEach(fwReport => {
                const fw = fwReport.framework;
                const complianceClass = fw.compliance_percentage >= 70 ? 'compliance-high' : 
                                      fw.compliance_percentage >= 40 ? 'compliance-medium' : 'compliance-low';
                
                // Calculate risk stats from controls
                let greenControls = 0, yellowControls = 0, redControls = 0;
                fwReport.controls.forEach(ctrl => {
                    if (ctrl.control.risk_status === 'green') greenControls++;
                    else if (ctrl.control.risk_status === 'yellow') yellowControls++;
                    else if (ctrl.control.risk_status === 'red') redControls++;
                });
                
                html += ` + "`" + `
                    <div class="framework">
                        <div class="framework-header" onclick="toggleFramework(this)">
                            <div>
                                <div class="framework-title">${fw.name}</div>
                                <div style="color: #666; margin-top: 5px;">${fwReport.controls.length} Controls</div>
                            </div>
                            <div class="compliance-badge ${complianceClass}">
                                ${fw.compliance_percentage.toFixed(1)}% Compliant
                            </div>
                        </div>
                        <div class="framework-body">
                            <div style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px;">
                                <div style="text-align: center; padding: 20px; background: #d4edda; border-radius: 8px;">
                                    <div style="font-size: 2em; font-weight: bold; color: #155724;">${greenControls}</div>
                                    <div style="color: #155724;">Low Risk</div>
                                </div>
                                <div style="text-align: center; padding: 20px; background: #fff3cd; border-radius: 8px;">
                                    <div style="font-size: 2em; font-weight: bold; color: #856404;">${yellowControls}</div>
                                    <div style="color: #856404;">Medium Risk</div>
                                </div>
                                <div style="text-align: center; padding: 20px; background: #f8d7da; border-radius: 8px;">
                                    <div style="font-size: 2em; font-weight: bold; color: #721c24;">${redControls}</div>
                                    <div style="color: #721c24;">High Risk</div>
                                </div>
                            </div>
                        </div>
                    </div>
                ` + "`" + `;
            });
            
            container.innerHTML = html;
        }

        function renderFrameworks() {
            const container = document.getElementById('frameworks-tab');
            let html = '<h2 style="margin-bottom: 20px;">Framework Controls</h2>';
            
            reportData.frameworks.forEach(fwReport => {
                const fw = fwReport.framework;
                html += ` + "`" + `<div class="framework expanded">
                    <div class="framework-header">
                        <div class="framework-title">${fw.name}</div>
                    </div>
                    <div class="framework-body">
                        <div class="controls-grid">` + "`" + `;
                
                fwReport.controls.forEach(ctrl => {
                    const control = ctrl.control;
                    const riskClass = control.risk_status === 'green' ? 'risk-green' : 
                                    control.risk_status === 'yellow' ? 'risk-yellow' : 'risk-red';
                    
                    html += ` + "`" + `
                        <div class="control-card" onclick="showControlDetail('${fw.id}', '${control.id}')">
                            <div class="control-id">
                                <span class="risk-indicator ${riskClass}"></span>
                                ${control.id}
                            </div>
                            <div class="control-title">${control.title}</div>
                            <div class="control-stats">
                                <span>📊 ${ctrl.evidence ? ctrl.evidence.length : 0} Evidence</span>
                                <span>⚠️ ${ctrl.findings ? ctrl.findings.length : 0} Findings</span>
                            </div>
                        </div>
                    ` + "`" + `;
                });
                
                html += '</div></div></div>';
            });
            
            container.innerHTML = html;
        }

        function renderFindings() {
            const container = document.getElementById('findings-tab');
            let html = '<h2 style="margin-bottom: 20px;">Compliance Findings</h2>';
            
            const allFindings = [];
            reportData.frameworks.forEach(fwReport => {
                const fw = fwReport.framework;
                fwReport.controls.forEach(ctrl => {
                    if (ctrl.findings) {
                        ctrl.findings.forEach(finding => {
                            allFindings.push({
                                framework: fw.name,
                                control: ctrl.control.id,
                                finding: finding
                            });
                        });
                    }
                });
            });
            
            if (allFindings.length === 0) {
                html += '<p style="text-align: center; color: #28a745; font-size: 1.2em; padding: 40px;">✅ No findings - All controls are compliant!</p>';
            } else {
                allFindings.forEach(item => {
                    const severityColor = item.finding.severity === 'critical' ? '#dc3545' :
                                        item.finding.severity === 'high' ? '#fd7e14' :
                                        item.finding.severity === 'medium' ? '#ffc107' : '#17a2b8';
                    
                    html += ` + "`" + `
                        <div class="evidence-item" style="border-left-color: ${severityColor};">
                            <div class="evidence-header">
                                <div>
                                    <strong>${item.framework} - ${item.control}</strong>
                                    <div style="color: #666; font-size: 0.9em; margin-top: 5px;">${item.finding.message}</div>
                                </div>
                                <span style="background: ${severityColor}; color: white; padding: 4px 12px; border-radius: 12px; font-size: 0.75em;">
                                    ${item.finding.severity.toUpperCase()}
                                </span>
                            </div>
                            ${item.finding.recommendation ? ` + "`<div style='margin-top: 10px; padding: 10px; background: white; border-radius: 4px;'><strong>💡 Recommendation:</strong> ${item.finding.recommendation}</div>`" + ` : ''}
                        </div>
                    ` + "`" + `;
                });
            }
            
            container.innerHTML = html;
        }

        function renderEvidence() {
            const container = document.getElementById('evidence-tab');
            let html = ` + "`" + `
                <h2 style="margin-bottom: 20px;">Evidence Analysis</h2>
                <div class="filters">
                    <button class="filter-btn active" onclick="filterEvidence('all')">All Evidence</button>
                    <button class="filter-btn" onclick="filterEvidence('ai')">🤖 AI-Enhanced</button>
                    <button class="filter-btn" onclick="filterEvidence('heuristic')">📋 Heuristic</button>
                </div>
                <div id="evidenceList"></div>
            ` + "`" + `;
            
            container.innerHTML = html;
            filterEvidence('all');
        }

        function filterEvidence(type) {
            currentFilter = type;
            document.querySelectorAll('.filters .filter-btn').forEach(btn => btn.classList.remove('active'));
            event.target.classList.add('active');
            
            const allEvidence = [];
            reportData.frameworks.forEach(fwReport => {
                const fw = fwReport.framework;
                fwReport.controls.forEach(ctrl => {
                    if (ctrl.evidence) {
                        ctrl.evidence.forEach(ev => {
                            if (type === 'all' || 
                                (type === 'ai' && ev.ai_analyzed) || 
                                (type === 'heuristic' && !ev.ai_analyzed)) {
                                allEvidence.push({
                                    framework: fw.name,
                                    control: ctrl.control.id,
                                    evidence: ev
                                });
                            }
                        });
                    }
                });
            });
            
            let html = '';
            allEvidence.slice(0, 50).forEach(item => {
                const ev = item.evidence;
                const aiClass = ev.ai_analyzed ? 'ai-enhanced' : '';
                
                html += ` + "`" + `
                    <div class="evidence-item ${aiClass}">
                        <div class="evidence-header">
                            <div>
                                <strong>${item.framework} - ${item.control}</strong>
                                ${ev.ai_analyzed ? '<span class="ai-badge">🤖 AI Enhanced</span>' : ''}
                            </div>
                            <span style="color: #667eea; font-weight: 600;">${Math.round(ev.confidence_score)}%</span>
                        </div>
                        <div style="font-size: 0.9em; color: #666; margin-top: 8px;">
                            ${ev.ai_analyzed ? ev.ai_justification : ev.reasoning}
                        </div>
                        <div class="confidence-bar">
                            <div class="confidence-fill" style="width: ${ev.confidence_score}%"></div>
                        </div>
                        <div style="font-size: 0.85em; color: #999; margin-top: 8px;">
                            ${ev.analysis_method}
                        </div>
                    </div>
                ` + "`" + `;
            });
            
            if (allEvidence.length > 50) {
                html += ` + "`<div style='text-align: center; padding: 20px; color: #666;'>Showing 50 of ${allEvidence.length} evidence entries</div>`" + `;
            }
            
            document.getElementById('evidenceList').innerHTML = html;
        }

        function countAIEvidence() {
            let count = 0;
            reportData.frameworks.forEach(fwReport => {
                fwReport.controls.forEach(ctrl => {
                    if (ctrl.evidence) {
                        ctrl.evidence.forEach(ev => {
                            if (ev.ai_analyzed) count++;
                        });
                    }
                });
            });
            return count;
        }

        function switchTab(tabName) {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            event.target.classList.add('active');
            
            document.querySelectorAll('.tab-content').forEach(tc => tc.style.display = 'none');
            document.getElementById(tabName + '-tab').style.display = 'block';
        }

        function toggleFramework(element) {
            element.parentElement.classList.toggle('expanded');
        }

        function showControlDetail(frameworkId, controlId) {
            const fwReport = reportData.frameworks.find(f => f.framework.id === frameworkId);
            const controlData = fwReport.controls.find(c => c.control.id === controlId);
            const control = controlData.control;
            
            const evidenceCount = controlData.evidence ? controlData.evidence.length : 0;
            const findingsCount = controlData.findings ? controlData.findings.length : 0;
            
            let html = ` + "`" + `
                <h2>${control.id}: ${control.title}</h2>
                <p style="margin: 15px 0; color: #666;">${control.description}</p>
                
                <h3 style="margin-top: 25px;">📊 Statistics</h3>
                <div style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px; margin: 15px 0;">
                    <div style="padding: 15px; background: #f8f9fa; border-radius: 8px;">
                        <div style="font-size: 1.5em; font-weight: bold; color: #667eea;">${evidenceCount}</div>
                        <div style="color: #666;">Evidence</div>
                    </div>
                    <div style="padding: 15px; background: #f8f9fa; border-radius: 8px;">
                        <div style="font-size: 1.5em; font-weight: bold; color: #667eea;">${findingsCount}</div>
                        <div style="color: #666;">Findings</div>
                    </div>
                    <div style="padding: 15px; background: #f8f9fa; border-radius: 8px;">
                        <div style="font-size: 1.5em; font-weight: bold; color: #667eea;">${Math.round(control.confidence_level)}</div>
                        <div style="color: #666;">Confidence</div>
                    </div>
                </div>
                
                <h3 style="margin-top: 25px;">🔍 Evidence (${evidenceCount})</h3>
            ` + "`" + `;
            
            if (controlData.evidence) {
                controlData.evidence.forEach(ev => {
                    const aiClass = ev.ai_analyzed ? 'ai-enhanced' : '';
                    html += ` + "`" + `
                        <div class="evidence-item ${aiClass}" style="margin-top: 10px;">
                            <div class="evidence-header">
                                <div>${ev.ai_analyzed ? '<span class="ai-badge">🤖 AI</span>' : '📋 Heuristic'}</div>
                                <span style="color: #667eea; font-weight: 600;">${Math.round(ev.confidence_score)}%</span>
                            </div>
                            <div style="font-size: 0.9em; color: #666; margin-top: 8px;">
                                ${ev.ai_analyzed ? ev.ai_justification : ev.reasoning}
                            </div>
                        </div>
                    ` + "`" + `;
                });
            }
            
            document.getElementById('modalBody').innerHTML = html;
            document.getElementById('detailModal').classList.add('active');
        }

        function closeModal() {
            document.getElementById('detailModal').classList.remove('active');
        }

        function handleSearch() {
            // Simple search implementation
            const query = document.getElementById('searchBox').value.toLowerCase();
            // TODO: Implement search filtering
        }

        // Initialize on load
        window.onload = init;
    </script>
</body>
</html>
`
