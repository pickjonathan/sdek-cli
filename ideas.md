feature suggestions :

** I would like to provide govrenence controles in diffrent formats, then ai should know ( using MCP's ) to fetch the right data and create analysed reports.
For example, I would like to provide CSP ( Swift ) controles ( csv, excel, plain text) it should create a policy excerpts like in the /Users/pickjonathan/WorkSpacePrivate/sdek-cli/internal/policy/excerpts.go then understand the data it needs to fetch for example from aws account, then analyse the findings according to the policy and provide a report.

** When using AI we should inject befor analysing the security complience with data on the complience we are analysing and the specific section of the complience.
example, we wat to analyse SOC2 so first we inject the AI with the SOC2 context and the specific excerpts we would like to check, then it should based on the information provided return a response with the finding.
At the second phase the AI given the complience ( SOC2 ) and the specific excerpt should alone conclude the data it needs to collect, get it using MCPS ( githu, AWS etc ) then perform the analysis.

** Evidence collection using AI web browser to login for example to AWS then return a print screen of the evidance.

** VS code extention to help configure the cli, run reports and view them, also connect MCP's.

** Installing the sdek tool via homebrew and other similer package management tools for other os's

** Web user interface with the CLI functionalities and enterprise features like audit exporting, user managment, sso, detailed reports ( extended reports ).

** AI agent for interaction mode based on findings, A GRC can live interact with the agent to get detailed information about the system.

** each one of the excerpts in internal/policy/excerpts.go should have also a defined prompt that needs to be injected to the agent incharge of anlayzing its section aand the prompt should be based on the governance section.

** Using multi agent framework to orchestrate a multi agent governance check, meaning there should be orchestrator that passes tasks to sub agents that gets injected with a governanace and a section in it they should master its checks, then return a summery report for the orchestrator to summerise and create a report out of.

** the cli should support more the stdout architecture, if for example a backend should use its functionalies it should run in http mode and provide http responses.

** The cli tool should be MCP compatible exposing its usages to major AI providers.