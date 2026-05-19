---
name: hospital-admin-guidelines
description: Coding conventions and patterns for the Hospital Admin Panel React project. Use when creating components, writing API calls, or structuring pages to ensure consistency across team members.
---

# Hospital Admin Panel — Coding Guidelines

## 1. Project Structure

Follow this folder structure strictly:

```
src/
├── api/          # 1 file per backend module (auth.js, flow.js, map.js...)
├── components/   # Shared reusable components (Layout/, GridCanvas/, etc.)
├── pages/        # 1 file per route/page (Dashboard.jsx, MapEditor.jsx...)
├── stores/       # Zustand stores (authStore.js)
├── hooks/        # Custom hooks (useAuth.js, useAutoRefresh.js)
├── utils/        # Pure helper functions
├── App.jsx       # Router configuration
└── main.jsx      # Entry point
```

## 2. API Layer Rules

**Always** use the centralized Axios client from `src/api/client.js`:

```js
// ✅ Correct
import api from './client';
export const fetchHeatmap = () => api.get('/flow/get_heatmap').then(r => r.data.data);

// ❌ Wrong - never use raw fetch or create new Axios instances
fetch('/api/flow/get_heatmap')
```

**Always** return `response.data.data` (unwrap the standard response wrapper).

API file naming: match backend module exactly (`flow.js`, `map.js`, `medical.js`, `device.js`, `sos.js`, `chat.js`, `engine.js`, `util.js`).

## 3. Data Fetching Rules

**Always** use TanStack Query. Never use `useEffect` + `useState` for API calls.

```js
// ✅ Correct
const { data, isLoading } = useQuery({
  queryKey: ['heatmap'],
  queryFn: fetchHeatmap,
});

// ❌ Wrong
const [data, setData] = useState(null);
useEffect(() => { fetchHeatmap().then(setData); }, []);
```

For mutations (POST/PATCH/DELETE):
```js
const queryClient = useQueryClient();
const mutation = useMutation({
  mutationFn: resolveObstacle,
  onSuccess: () => {
    message.success('Done');
    queryClient.invalidateQueries({ queryKey: ['obstacles'] });
  },
});
```

## 4. Component Rules

- Use **function components** only (no class components)
- Use **Ant Design** components for all UI (Table, Form, Modal, Card, Button...)
- Do NOT install additional UI libraries (no Material UI, no Chakra)
- 1 file = 1 default export component
- Component name = PascalCase, file name = PascalCase

```js
// ✅ Correct
export default function FlowMonitor() { ... }

// ❌ Wrong
export const flowMonitor = () => { ... }
```

## 5. Page Layout Pattern

Every page must follow this structure:

```jsx
import { Card, Row, Col, Typography } from 'antd';
const { Title } = Typography;

export default function PageName() {
  return (
    <>
      <Title level={4}>Page Title</Title>
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card title="Section Name">
            {/* content */}
          </Card>
        </Col>
      </Row>
    </>
  );
}
```

## 6. State Management

- **Server state** (API data) → TanStack Query (never Zustand)
- **Client state** (auth, sidebar, theme) → Zustand
- **Form state** → Ant Design Form (useForm hook)
- **URL state** (filters, pagination) → React Router searchParams

## 7. Error Handling

```js
// API errors are handled globally by Axios interceptor (401 → redirect to login)
// For page-level errors, use TanStack Query's error state:
const { data, error, isLoading } = useQuery({...});

if (error) return <Result status="error" title="Failed to load" />;
if (isLoading) return <Spin />;
```

## 8. Auto-refresh Pattern

For realtime data (heatmap, SOS, queue), use `refetchInterval`:

```js
useQuery({
  queryKey: ['heatmap'],
  queryFn: fetchHeatmap,
  refetchInterval: 5000,  // 5 seconds
});
```

Do NOT use `setInterval` + manual fetch.

## 9. Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| Component | PascalCase | `FlowMonitor.jsx` |
| API file | camelCase | `flow.js` |
| Hook | camelCase with `use` prefix | `useAuth.js` |
| Store | camelCase with `Store` suffix | `authStore.js` |
| CSS | CSS Modules or Ant Design inline | `FlowMonitor.module.css` |
| Constants | UPPER_SNAKE | `API_BASE_URL` |

## 10. Git Workflow

```
main ← develop ← feature/xxx
```

- Never push directly to `main` or `develop`
- Always create feature branch: `feature/<page-name>`
- Commit messages: `feat:`, `fix:`, `docs:`, `style:`, `refactor:`
- PR must be reviewed by at least 1 person before merge
