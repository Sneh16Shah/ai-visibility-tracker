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

---

### 🎨 Code Quality: Frontend Styling Refactor

**Problem:** Current inline Tailwind classes are hard to read and maintain:
```jsx
// ❌ Current (messy)
<button className="px-6 py-3 bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white font-medium rounded-xl shadow-lg shadow-indigo-500/25 transition-all duration-300">
```

**Solution:** Adopt industry-standard CSS organization like big companies:

| Approach | Used By | Description |
|----------|---------|-------------|
| **CSS Modules** | Meta, Vercel | `.module.css` files scoped per component |
| **Styled Components** | Airbnb, Spotify | CSS-in-JS with tagged template literals |
| **Design Tokens** | Salesforce, Adobe | `tokens.css` with CSS variables |
| **BEM + Utility** | GitHub, Stripe | Semantic class names + utilities |

**Recommended Approach for this project:**

1. **Create Design Tokens** (`styles/tokens.css`)
   ```css
   :root {
     --btn-primary-bg: linear-gradient(to right, #6366f1, #9333ea);
     --btn-primary-hover: linear-gradient(to right, #4f46e5, #7e22ce);
     --radius-lg: 0.75rem;
     --shadow-primary: 0 10px 15px -3px rgb(99 102 241 / 0.25);
   }
   ```

2. **Create Component Classes** (`styles/components.css`)
   ```css
   .btn-primary {
     background: var(--btn-primary-bg);
     color: white;
     padding: 0.75rem 1.5rem;
     border-radius: var(--radius-lg);
     box-shadow: var(--shadow-primary);
   }
   .btn-primary:hover {
     background: var(--btn-primary-hover);
   }
   ```

3. **Use Semantic Classes in JSX**
   ```jsx
   // ✅ Clean & readable
   <button className="btn-primary">Run Analysis</button>
   ```

**Tasks:**
- [ ] Create `styles/tokens.css` with design variables
- [ ] Create `styles/components.css` with reusable component classes
- [ ] Refactor Dashboard.jsx to use semantic classes
- [ ] Refactor RunAnalysis.jsx to use semantic classes
- [ ] Refactor Brands.jsx to use semantic classes
- [ ] Document styling conventions in README


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

free api
openrouter