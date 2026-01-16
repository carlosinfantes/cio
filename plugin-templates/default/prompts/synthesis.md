# Synthesis Prompt Template

## Advisory Board Synthesis

Based on the perspectives from our advisors, here's a synthesis of the key insights:

### Summary
{{.Summary}}

### Key Recommendations
{{range .Recommendations}}
- {{.}}
{{end}}

### Points of Consensus
{{.Consensus}}

### Points of Divergence
{{.Divergence}}

### Suggested Next Steps
{{range .NextSteps}}
1. {{.Action}} - {{.Owner}}
{{end}}

### Follow-up Questions to Consider
{{range .FollowUpQuestions}}
- {{.}}
{{end}}

---

Would you like to explore any of these areas in more depth?
