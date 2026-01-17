## All Tasks:

### 🔥 Priority (Current Sprint)
1. Login/Signup - (add register/login options and remove demo part)
2. Dashboard → Run Analysis redirect with correct brand ID
3. Analysis completion modal → redirect to Dashboard
4. Analysis - Custom questions/prompts support
5. Analysis - Fix rate limit UX on run button
6. Dashboard - Show API call count during analysis
7. Brands - Cascade delete all related data

### ⭐ High Value Features (Next Sprint)
8. **Multi-AI Comparison** - Run same prompts on GPT, Gemini, Claude & compare results
9. **Export Reports** - PDF/CSV export of visibility metrics & trends
10. **Email Alerts** - Notify when visibility score drops below threshold
11. **Scheduled Analysis** - Auto-run daily/weekly analysis (cron jobs)
12. **Competitor Deep Dive** - Analyze WHY competitors rank better

### 💡 Nice-to-Have Features
13. **Historical Trends** - Long-term visibility graphs (30/60/90 days)
14. **GEO Recommendations** - Actionable tips to improve AI visibility
15. **API Access** - REST API for external integrations
16. **Team/Agency Mode** - Multiple users, brand groups, permissions
17. **White-label** - Custom branding for agencies
18. **Mobile Responsive** - Fully optimized for mobile viewing
19. **Dark/Light Mode Toggle** - Theme switching


1. AI Visibility Enhancement Feature

### Feature Idea
Allow users to take actions that **increase their brand's visibility in AI models**.

### Is it possible?
**Yes!** This is called **Generative Engine Optimization (GEO)** - optimizing content to appear more frequently in AI-generated responses.

### Strategies that could be offered:

| Strategy | Description | How it works |
|----------|-------------|--------------|
| **Structured Data** | Add schema markup to website | AI models parse structured data better |
| **Authoritative Content** | Create expert-level content | AI prioritizes high-quality, cited sources |
| **Wikipedia Presence** | Get brand on Wikipedia | Major AI training data source |
| **Knowledge Panels** | Google Knowledge Graph | AI models often reference this |
| **FAQ Content** | Answer common questions | Matches how users prompt AI |
| **Industry Publications** | Get featured in trade sites | Increases authoritative mentions |
| **Reviews & Testimonials** | Boost on review platforms | AI cites user reviews |

### Proposed Feature Flow
1. User runs analysis → sees low visibility score
2. Dashboard shows "Improve Visibility" recommendations
3. Clicking shows personalized action items based on:
   - Current visibility gaps
   - Competitor analysis
   - Industry best practices
4. Track improvement over time

### Implementation Approach
- Add "Recommendations" section to Dashboard
- Analyze AI responses to identify WHY brand isn't mentioned
- Generate actionable suggestions based on gaps
