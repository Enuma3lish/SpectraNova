# FenzVideo Frontend Architecture

## Tech Stack (100% Open Source)

| Category         | Technology                                                                                        | License     | Description                                            |
| ---------------- | ------------------------------------------------------------------------------------------------- | ----------- | ------------------------------------------------------ |
| Framework        | [Vue 3](https://vuejs.org/) (Composition API + `<script setup>`)                                  | MIT         | Progressive JavaScript framework                       |
| Build Tool       | [Vite](https://vite.dev/) 7.3+                                                                    | MIT         | Fast build tool & dev server                           |
| Routing          | [Vue Router](https://router.vuejs.org/) 4                                                         | MIT         | Official Vue routing                                   |
| State Management | [Pinia](https://pinia.vuejs.org/) 3                                                               | MIT         | Official Vue state management                          |
| HTTP Client      | [Axios](https://axios-http.com/) 1.13+                                                            | MIT         | Promise-based HTTP client                              |
| UI Framework     | [Element Plus](https://element-plus.org/) 2.13+                                                   | MIT         | Vue 3 component library                                |
| CSS Utility      | [Tailwind CSS](https://tailwindcss.com/) 3.4                                                      | MIT         | Utility-first CSS framework (supplements Element Plus) |
| Video Player     | [Video.js](https://videojs.com/) 8.23+                                                            | Apache-2.0  | HTML5 video player                                     |
| i18n             | [Vue I18n](https://vue-i18n.intlify.dev/) 10                                                      | MIT         | Internationalization (zh-TW / en)                      |
| Auth             | JWT (stored in localStorage)                                                                      | —           | Token-based authentication                             |
| Icons            | [Iconify](https://iconify.design/) / [@iconify/vue](https://github.com/iconify/iconify) 5        | MIT         | Open-source icon sets                                  |
| Linting          | [ESLint](https://eslint.org/)                                                                     | MIT         | Code quality                                           |
| Language         | [TypeScript](https://www.typescriptlang.org/) ~5.9                                                | Apache-2.0  | Type-safe JavaScript                                   |

### Planned Dependencies (Not Yet Installed)

| Technology | Phase | Purpose |
|------------|-------|---------|
| [ECharts](https://echarts.apache.org/) via [vue-echarts](https://github.com/ecomfe/vue-echarts) | Phase 4 | Dashboard analytics charts |
| [VeeValidate](https://vee-validate.logaretm.com/) + [Yup](https://github.com/jquense/yup) | Phase 4 | Schema-based form validation |
| [Paddle.js](https://developer.paddle.com/paddlejs/overview) | Phase 4 | Client-side payment checkout |
| [Vitest](https://vitest.dev/) + [Vue Test Utils](https://test-utils.vuejs.org/) | Future | Unit & component testing |
| [Playwright](https://playwright.dev/) | Future | End-to-end browser testing |

---

## Directory Structure

```
frontend/
├── public/
│   └── favicon.ico
├── src/
│   ├── api/                    # API layer (Axios instances & endpoints)
│   │   ├── index.ts            # ✅ Axios instance with interceptors
│   │   ├── auth.ts             # ✅ Login / Register
│   │   ├── video.ts            # ✅ Video CRUD, recommended
│   │   ├── channel.ts          # ✅ Channel info, subscribe/unsubscribe
│   │   ├── category.ts         # ✅ Categories
│   │   ├── tag.ts              # ✅ Tags, get/set user tags
│   │   ├── search.ts           # ✅ Search with filters
│   │   └── admin.ts            # ✅ Admin user/video/tag management
│   │
│   ├── assets/                 # Static assets
│   │   ├── styles/
│   │   │   ├── variables.scss
│   │   │   └── global.css
│   │   ├── images/
│   │   └── vue.svg
│   │
│   ├── components/             # Reusable UI components
│   │   ├── common/
│   │   │   ├── AppHeader.vue          # ✅ Top navigation (search + login + admin link)
│   │   │   ├── AppSidebar.vue         # ✅ Categories + TagSelector
│   │   │   ├── SearchBar.vue          # ✅ Global search input
│   │   │   ├── VideoCard.vue          # ✅ Video thumbnail card
│   │   │   ├── VideoGrid.vue          # ✅ Grid layout of VideoCards
│   │   │   ├── Pagination.vue         # ✅ Pagination control
│   │   │   ├── ConfirmDialog.vue      # ✅ Reusable confirmation modal
│   │   │   └── LoadingSpinner.vue     # ✅ Loading indicator
│   │   │
│   │   ├── auth/
│   │   │   ├── LoginForm.vue          # ✅ Login form
│   │   │   └── RegisterForm.vue       # ✅ Register form
│   │   │
│   │   └── tag/
│   │       └── TagSelector.vue        # ✅ Pick up to 5 tags (chips)
│   │
│   ├── layouts/                # Layout wrappers
│   │   ├── DefaultLayout.vue   # ✅ Header + Sidebar + Main content
│   │   ├── AuthLayout.vue      # ✅ Centered card (login/register)
│   │   └── AdminLayout.vue     # ✅ Admin sidebar + content area
│   │
│   ├── router/
│   │   └── index.ts            # ✅ Route definitions + guards
│   │
│   ├── stores/                 # Pinia stores
│   │   ├── authStore.ts        # ✅ User auth state, JWT token, role
│   │   ├── videoStore.ts       # ✅ Recommended videos, current video
│   │   ├── categoryStore.ts    # ✅ Category list
│   │   ├── tagStore.ts         # ✅ Available tags, user selected tags, sessionId
│   │   ├── searchStore.ts      # ✅ Search query, filters, results
│   │   └── adminStore.ts       # ✅ Admin user/video/tag management
│   │
│   ├── types/                  # TypeScript type definitions
│   │   ├── user.ts             # ✅ User type (with role, isHidden)
│   │   ├── video.ts            # ✅ Video type (with tags, accessTier)
│   │   ├── channel.ts          # ✅ Channel type
│   │   ├── category.ts         # ✅ Category type
│   │   ├── tag.ts              # ✅ Tag type
│   │   ├── search.ts           # ✅ SearchFilters type
│   │   ├── api.ts              # ✅ API response types
│   │   └── router.d.ts         # ✅ Vue Router meta extensions
│   │
│   ├── utils/                  # Utility functions
│   │   ├── formatDate.ts       # ✅ Date formatting
│   │   ├── formatDuration.ts   # ✅ Duration formatting (MM:SS / H:MM:SS)
│   │   └── formatViews.ts      # ✅ View count formatting (1.2K, 1.5M)
│   │
│   ├── views/                  # Page-level components (bound to routes)
│   │   ├── LoginView.vue              # ✅ Login + Register tabs
│   │   ├── HomeView.vue               # ✅ Tag-based recommended videos
│   │   ├── SearchResultsView.vue      # ✅ Search + advanced filters
│   │   ├── CategoryView.vue           # ✅ Videos by category
│   │   ├── ChannelView.vue            # ✅ Channel profile + subscribe
│   │   ├── VideoView.vue              # ✅ Video player + metadata
│   │   └── admin/
│   │       ├── AdminView.vue          # ✅ Admin main wrapper
│   │       ├── AdminUsersView.vue     # ✅ User management (list + delete)
│   │       └── AdminTagsView.vue      # ✅ Tag CRUD (via dialog form)
│   │
│   ├── App.vue                 # ✅ Dynamic layout selection
│   ├── main.ts                 # ✅ App entry point
│   └── style.css               # ✅ Global styles
│
├── .env                        # VITE_API_BASE_URL, VITE_APP_TITLE
├── .dockerignore               # Docker build excludes
├── Dockerfile                  # Multi-stage build: Node 20 → Nginx Alpine
├── nginx.conf                  # Nginx config: SPA routing, /api/ proxy, /fenzvideo/ MinIO proxy
├── index.html
├── package.json
├── tsconfig.json
├── tsconfig.app.json
├── tsconfig.node.json
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
└── eslint.config.js
```

### Planned Components (Not Yet Implemented)

| Component | Phase | Purpose |
|-----------|-------|---------|
| `TagPicker.vue` | Phase 4 | Inline tag picker for video upload form |
| `VideoUploadForm.vue` | Phase 4 | Upload form with category/tag/access tier |
| `VideoEditForm.vue` | Phase 4 | Edit existing video |
| `VideoDonateDialog.vue` | Phase 4 | Paddle.js donation checkout |
| `VideoFilterPanel.vue` | Phase 4 | Enhanced search filter sidebar |
| `ChannelBanner.vue` | Phase 4 | Channel header with avatar & name |
| `MembershipDialog.vue` | Phase 4 | Join/Leave membership modal |
| `ChannelVideoList.vue` | Phase 4 | Channel's video listing |
| `DashboardVideoList.vue` | Phase 4 | Uploaded videos management |
| `MembershipFeeForm.vue` | Phase 4 | Set monthly fee |
| `AnalyticsCharts.vue` | Phase 4 | ECharts dashboard wrapper |
| `AccountSettings.vue` | Phase 5 | Display name, password, hide/delete |
| `DashboardLayout.vue` | Phase 4 | Dashboard sidebar + content |

---

## Routing

```ts
// router/index.ts
const routes = [
  {
    path: "/login",
    name: "Login",
    component: LoginView,
    meta: { layout: "auth", guestOnly: true },
  },
  {
    path: "/",
    name: "Home",
    component: HomeView,
    meta: { layout: "default" },
  },
  {
    path: "/search",
    name: "Search",
    component: SearchResultsView,
    meta: { layout: "default" },
  },
  {
    path: "/category/:id",
    name: "Category",
    component: CategoryView,
    meta: { layout: "default" },
  },
  {
    path: "/channel/:id",
    name: "Channel",
    component: ChannelView,
    meta: { layout: "default" },
  },
  {
    path: "/video/:id",
    name: "Video",
    component: VideoView,
    meta: { layout: "default" },
  },
  {
    path: "/admin",
    name: "Admin",
    component: AdminView,
    meta: { layout: "admin", requiresAuth: true, requiresAdmin: true },
    children: [
      { path: "", redirect: { name: "AdminUsers" } },
      { path: "users", name: "AdminUsers", component: AdminUsersView },
      { path: "tags", name: "AdminTags", component: AdminTagsView },
    ],
  },
];
```

### Navigation Guards

```ts
// router/index.ts (beforeEach)
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore();

  // Redirect logged-in users away from login page
  if (to.meta.guestOnly && authStore.isLoggedIn) {
    return next({ name: "Home" });
  }

  // Protect authenticated routes
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    return next({ name: "Login", query: { redirect: to.fullPath } });
  }

  // Protect admin routes — must have role === 'admin'
  if (to.meta.requiresAdmin && authStore.user?.role !== "admin") {
    return next({ name: "Home" });
  }

  next();
});
```

---

## State Management (Pinia Stores)

### authStore

```ts
interface AuthState {
  user: User | null;
  token: string | null;       // from localStorage
  refreshToken: string | null; // from localStorage
}

// Computed:
//   isLoggedIn: !!token
//   isAdmin: user?.role === 'admin'

// Actions: login(), register(), logout(), refreshAuthToken()
```

### videoStore

```ts
interface VideoState {
  recommendedVideos: Video[];
  currentVideo: Video | null;
  totalRecommended: number;
}

// Actions: fetchRecommended(page, pageSize, sessionId), fetchVideo(id)
```

### tagStore

```ts
interface TagState {
  allTags: Tag[];
  selectedTags: Tag[];
  sessionId: string;      // auto-generated UUID, persisted in localStorage
}

// Actions: fetchAllTags(), fetchMyTags(), setMyTags(tagIds[])
// Guest flow: if not logged in, uses sessionId (UUID in localStorage)
```

### searchStore

```ts
interface SearchState {
  query: string;
  filters: SearchFilters;
  results: Video[];
  totalCount: number;
  page: number;
  pageSize: number;
}

interface SearchFilters {
  query?: string;
  category_id?: number;
  min_duration?: number;
  max_duration?: number;
  start_date?: string;
  end_date?: string;
  sort_by?: string;        // views_desc, views_asc, date_desc, date_asc
  access_type?: string;    // public, member
}

// Actions: search(), resetFilters()
```

### categoryStore

```ts
interface CategoryState {
  categories: Category[];
}

// Actions: fetchCategories()
```

### adminStore

```ts
interface AdminState {
  users: User[];
  totalUsers: number;
  videos: AdminVideo[];
  totalVideos: number;
  tags: Tag[];
}

interface AdminVideo {
  id: number;
  title: string;
  username: string;
  userId: number;
  categoryName: string;
  accessTier: number;
  isPublished: boolean;
  isHidden: boolean;
  viewsMember: number;
  viewsNonMember: number;
  createdAt: string;
}

// Actions:
//   Users: fetchUsers(page, pageSize), deleteUser(id)
//   Videos: fetchVideos(page, pageSize), deleteVideo(id)
//   Tags: fetchTags(), createTag(name, slug), updateTag(id, name, slug), deleteTag(id)
```

### Planned Stores (Not Yet Implemented)

| Store | Phase | Purpose |
|-------|-------|---------|
| `channelStore` | Phase 4 | Channel data, membership state |
| `dashboardStore` | Phase 4 | Dashboard analytics data |
| `donationStore` | Phase 4 | Donation sent/received lists |

---

## API Layer

### Axios Instance

```ts
// api/index.ts
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL, // /api/v1
  timeout: 30000,
  headers: { "Content-Type": "application/json" },
});

// Request interceptor — attach JWT from localStorage
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// Response interceptor — handle 401
apiClient.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem("token");
      localStorage.removeItem("refreshToken");
      router.push("/login");
    }
    return Promise.reject(err);
  },
);
```

### Endpoint Modules

| Module         | Endpoints                                                                                                                                    |
| -------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `auth.ts`      | `POST /auth/login`, `POST /auth/register`, `POST /auth/refresh`                                                                              |
| `video.ts`     | `GET /recommended`, `GET /videos/:id`, `POST /videos`, `PUT /videos/:id`, `DELETE /videos/:id`, `PATCH /videos/:id/publish`                   |
| `channel.ts`   | `GET /channels/:id`, `POST /channels/:id/subscribe`, `DELETE /channels/:id/subscribe`                                                         |
| `category.ts`  | `GET /categories`                                                                                                                            |
| `tag.ts`       | `GET /tags`, `GET /tags/my?session_id=...`, `PUT /tags/my`                                                                                   |
| `search.ts`    | `GET /search?query=...&category_id=...&min_duration=...&max_duration=...&sort_by=...&access_type=...`                                        |
| `admin.ts`     | `GET /admin/users`, `DELETE /admin/users/:id`, `GET /admin/videos`, `DELETE /admin/videos/:id`, `POST /admin/tags`, `PUT /admin/tags/:id`, `DELETE /admin/tags/:id` |

> All endpoints are prefixed with `/api/v1` by the Axios base URL.

---

## Key Component Interactions

```
┌─────────────────────────────────────────────────────┐
│                    App.vue                          │
│  ┌───────────────────────────────────────────────┐  │
│  │  Layout (Default / Auth / Admin)              │  │
│  │  ┌─────────────┐  ┌───────────────────────┐   │  │
│  │  │  AppHeader   │  │   <router-view />     │   │  │
│  │  │  (SearchBar) │  │   (Page Components)   │   │  │
│  │  │  (Admin link)│  │                       │   │  │
│  │  └─────────────┘  └───────────────────────┘   │  │
│  │  ┌─────────────┐                              │  │
│  │  │  AppSidebar  │                              │  │
│  │  │  (Categories)│                              │  │
│  │  │  (TagSelect) │  ← pick up to 5 tags        │  │
│  │  └─────────────┘                              │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Page → Component Mapping

| Page (View)              | Key Components Used                                   |
| ------------------------ | ----------------------------------------------------- |
| `LoginView`              | `LoginForm`, `RegisterForm` (tabbed)                  |
| `HomeView`               | `VideoGrid`, `VideoCard`, `Pagination`                |
| `SearchResultsView`      | `VideoGrid`, `VideoCard`, `Pagination` (inline filters) |
| `CategoryView`           | `VideoGrid`, `VideoCard`, `Pagination`                |
| `ChannelView`            | Channel info card, subscribe/unsubscribe button       |
| `VideoView`              | HTML5 video player, video info, channel link          |
| `AdminUsersView`         | El-Table, `ConfirmDialog`, `Pagination`               |
| `AdminTagsView`          | El-Table, El-Dialog form, `ConfirmDialog`             |

---

## Auth Flow

```
  User                    Frontend                     Backend
   │                         │                            │
   │── Enter credentials ──▶ │                            │
   │                         │── POST /auth/login ──────▶ │
   │                         │◀── { token, user } ───── │
   │                         │── Save to localStorage ─▶  │
   │◀── Redirect to Home ── │                            │
   │                         │                            │
   │── Access Admin ──────▶  │                            │
   │                         │── Guard checks token ──▶   │
   │                         │   & role === 'admin'       │
   │                         │── GET /admin/** ──────────▶ │
   │                         │   (Authorization: Bearer)  │
   │                         │◀── 200 Data ────────────── │
   │◀── Render Admin ──────  │                            │
```

---

## Environment Variables

```env
# .env
VITE_API_BASE_URL=/api/v1
VITE_APP_TITLE=FenzVideo
```

---

## Build & Dev

```bash
# Development
npm run dev          # Vite dev server at localhost:5173
                     # Proxy: /api → localhost:8000
                     # Proxy: /fenzvideo → localhost:9100 (MinIO)

# Production build
npm run build        # Output to dist/

# Preview production build
npm run preview

# Lint
npm run lint         # ESLint
```

---

## Docker Deployment

### Dockerfile (Multi-Stage Build)

```dockerfile
# Stage 1: Build the Vue app
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 2: Serve with Nginx
FROM nginx:stable-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Nginx Configuration (`nginx.conf`)

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # Use Docker's internal DNS resolver
    resolver 127.0.0.11 valid=10s;

    # Gzip compression
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml;

    # API proxy to backend
    location /api/ {
        set $backend http://fenzvideo-backend:8000;
        proxy_pass $backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        client_max_body_size 500M;
    }

    # MinIO proxy for video/thumbnail files
    location /fenzvideo/ {
        set $minio http://fenzvideo-minio:9000;
        proxy_pass $minio;
        proxy_set_header Host $host;
        proxy_buffering off;
        proxy_request_buffering off;
    }

    # SPA fallback — all non-file routes serve index.html
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Cache static assets (exclude MinIO-proxied paths)
    location ~* ^(?!/fenzvideo/).*\.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

**Key design decisions:**

| Decision | Rationale |
|---|---|
| Docker DNS resolver `127.0.0.11` | Enables variable-based `proxy_pass`, resolving container names at request time instead of startup |
| Variable-based `proxy_pass` | Prevents Nginx from crash-looping if backend/minio containers restart |
| `/fenzvideo/` proxy to MinIO | Serves video and thumbnail files from MinIO through the same origin, avoiding CORS issues |
| Static asset regex excluding `/fenzvideo/` | `^(?!/fenzvideo/)` prevents the cache regex from intercepting MinIO-proxied files (e.g., `/fenzvideo/thumbnails/thumb_1.jpg`) |
| `client_max_body_size 500M` | Supports large video file uploads through the API proxy |
| `proxy_buffering off` for MinIO | Allows streaming of large video files without buffering in Nginx |

---

## Key Dependencies (`package.json`)

```json
{
  "dependencies": {
    "vue": "^3.5.25",
    "vue-router": "^4.6.4",
    "pinia": "^3.0.4",
    "axios": "^1.13.5",
    "element-plus": "^2.13.2",
    "video.js": "^8.23.7",
    "vue-i18n": "^10.0.8",
    "@iconify/vue": "^5.0.0"
  },
  "devDependencies": {
    "vite": "^7.3.1",
    "typescript": "~5.9.3",
    "@vitejs/plugin-vue": "^5.2",
    "tailwindcss": "^3.4.19",
    "postcss": "^8.5",
    "autoprefixer": "^10.4",
    "eslint": "^9.22"
  }
}
```

---

## Planned Frontend Features (Future Phases)

### Phase 4 — Monetization Pages

- **DashboardView** — Video management + analytics (ECharts)
- **DashboardUploadView** — Video upload with category/tag/access tier
- **DashboardAnalyticsView** — Views, member count, revenue charts
- **DashboardDonationsView** — Sent & received donations
- **DashboardSettingsView** — Membership fee, account settings
- **VideoUploadForm** — Upload form component
- **VideoDonateDialog** — Paddle.js checkout overlay
- **MembershipDialog** — Join/Leave membership modal
- Paddle.js integration for payment

### Phase 5 — Advanced Features

- **Notification bell** component
- **WebSocket realtime store** for authenticated live events
- **Creator live alerts** while online: likes from viewers and moderation notices from admins
- **User profile** / self-service pages (hide/delete account)
- **AnalyticsCharts** enhancements
- E2E testing with Playwright
- Unit testing with Vitest

### Planned Real-Time Client Flow

- `authStore` provides the JWT used for the WebSocket handshake after login
- A new `realtimeStore` owns one browser WebSocket connection per tab and reconnects when tokens refresh
- `VideoView` emits the like action through HTTP; the creator does not receive the alert directly from the viewer browser
- The creator dashboard listens for server-pushed events:
  - `video_liked` shows a transient toast plus an inbox entry
  - `moderation_removed` shows a high-priority warning and navigates to the affected video record when available
- If the creator is offline, the frontend falls back to polling `NotificationService` and unread counts when the session resumes

---

## Open-Source Alternatives Reference

| Component       | Current           | Alternative (also open source) |
| --------------- | ----------------- | ------------------------------ |
| UI Framework    | Element Plus      | Naive UI, PrimeVue, Vuetify 3  |
| CSS             | Tailwind CSS      | UnoCSS, Windi CSS              |
| Video Player    | Video.js          | Plyr, Shaka Player, hls.js     |
| Icons           | Iconify           | Lucide, Heroicons              |
| Build Tool      | Vite              | Rspack, Farm                   |
