# FenzVideo Frontend Architecture

## Tech Stack (100% Open Source)

| Category         | Technology                                                                                        | License     | Description                                            |
| ---------------- | ------------------------------------------------------------------------------------------------- | ----------- | ------------------------------------------------------ |
| Framework        | [Vue 3](https://vuejs.org/) (Composition API + `<script setup>`)                                  | MIT         | Progressive JavaScript framework                       |
| Build Tool       | [Vite](https://vite.dev/)                                                                         | MIT         | Fast build tool & dev server                           |
| Routing          | [Vue Router](https://router.vuejs.org/) 4                                                         | MIT         | Official Vue routing                                   |
| State Management | [Pinia](https://pinia.vuejs.org/)                                                                 | MIT         | Official Vue state management                          |
| HTTP Client      | [Axios](https://axios-http.com/)                                                                  | MIT         | Promise-based HTTP client                              |
| UI Framework     | [Element Plus](https://element-plus.org/)                                                         | MIT         | Vue 3 component library                                |
| CSS Utility      | [Tailwind CSS](https://tailwindcss.com/)                                                          | MIT         | Utility-first CSS framework (supplements Element Plus) |
| Video Player     | [Video.js](https://videojs.com/)                                                                  | Apache-2.0  | HTML5 video player                                     |
| Charts           | [ECharts](https://echarts.apache.org/) (via [vue-echarts](https://github.com/ecomfe/vue-echarts)) | Apache-2.0  | Charting library by Apache                             |
| Form Validation  | [VeeValidate](https://vee-validate.logaretm.com/) + [Yup](https://github.com/jquense/yup)         | MIT         | Schema-based form validation                           |
| i18n             | [Vue I18n](https://vue-i18n.intlify.dev/)                                                         | MIT         | Internationalization (zh-TW / en)                      |
| Auth             | JWT (stored in httpOnly cookie or localStorage)                                                   | —           | Token-based authentication                             |
| Testing          | [Vitest](https://vitest.dev/) + [Vue Test Utils](https://test-utils.vuejs.org/)                   | MIT         | Unit & component testing                               |
| E2E Testing      | [Playwright](https://playwright.dev/)                                                             | Apache-2.0  | End-to-end browser testing                             |
| Payment          | [Paddle.js](https://developer.paddle.com/paddlejs/overview) (`@paddle/paddle-js`)                 | Proprietary | Client-side checkout overlay (sandbox)                 |
| Linting          | [ESLint](https://eslint.org/) + [Prettier](https://prettier.io/)                                  | MIT         | Code quality & formatting                              |
| Icons            | [Iconify](https://iconify.design/) / [unplugin-icons](https://github.com/unplugin/unplugin-icons) | MIT         | Open-source icon sets                                  |

---

## Directory Structure

```
frontend/
├── public/
│   └── favicon.ico
├── src/
│   ├── api/                    # API layer (Axios instances & endpoints)
│   │   ├── index.ts            # Axios instance with interceptors
│   │   ├── auth.ts             # Login / Register / Logout
│   │   ├── video.ts            # Video CRUD, upload, search
│   │   ├── channel.ts          # Channel info, membership
│   │   ├── category.ts         # Categories
│   │   ├── dashboard.ts        # Dashboard analytics, settings
│   │   ├── user.ts             # User profile, password change, self-delete
│   │   ├── tag.ts              # Tag list, get/set user tags
│   │   ├── donation.ts         # Create donation, list sent/received donations
│   │   └── admin.ts            # Admin user CRUD, tag CRUD, hide/restore/delete
│   │
│   ├── assets/                 # Static assets (images, fonts, global CSS)
│   │   ├── styles/
│   │   │   ├── variables.scss
│   │   │   └── global.scss
│   │   └── images/
│   │
│   ├── components/             # Reusable UI components
│   │   ├── common/
│   │   │   ├── AppHeader.vue          # Top navigation bar (admin link if role=admin)
│   │   │   ├── AppFooter.vue
│   │   │   ├── AppSidebar.vue         # Categories + TagSelector
│   │   │   ├── SearchBar.vue          # Global search input
│   │   │   ├── VideoCard.vue          # Video thumbnail card (shows hidden badge)
│   │   │   ├── VideoGrid.vue          # Grid layout of VideoCards
│   │   │   ├── Pagination.vue
│   │   │   ├── ConfirmDialog.vue      # Reusable confirmation modal
│   │   │   └── LoadingSpinner.vue
│   │   │
│   │   ├── auth/
│   │   │   ├── LoginForm.vue
│   │   │   └── RegisterForm.vue
│   │   │
│   │   ├── tag/
│   │   │   ├── TagSelector.vue        # Pick up to 5 tags (chips + checkboxes)
│   │   │   └── TagPicker.vue          # Inline tag picker for video upload form
│   │   │
│   │   ├── video/
│   │   │   ├── VideoPlayer.vue        # Video.js wrapper
│   │   │   ├── VideoInfo.vue          # Title, views, date, category, tags, description
│   │   │   ├── VideoUploadForm.vue    # Upload form (title, desc, category, tags, membership)
│   │   │   ├── VideoEditForm.vue      # Edit existing video
│   │   │   ├── VideoDonateDialog.vue  # Donate to creator modal (Paddle.js checkout, video-level)
│   │   │   └── VideoFilterPanel.vue   # Search filter sidebar
│   │   │
│   │   ├── channel/
│   │   │   ├── ChannelBanner.vue      # Channel header with avatar & name
│   │   │   ├── MembershipDialog.vue   # Join / Leave membership modal
│   │   │   └── ChannelVideoList.vue
│   │   │
│   │   ├── dashboard/
│   │   │   ├── DashboardVideoList.vue # Uploaded videos management table
│   │   │   ├── MembershipFeeForm.vue  # Set monthly fee
│   │   │   ├── AnalyticsCharts.vue    # ECharts wrapper
│   │   │   ├── AccountSettings.vue    # Display name, password, hide/delete account
│   │   │   ├── DonationHistoryTable.vue # Sent & received donations table
│   │   │   └── charts/
│   │   │       ├── TotalViewsChart.vue
│   │   │       ├── ViewsRankingChart.vue
│   │   │       ├── MemberCountChart.vue
│   │   │       ├── MemberRatioChart.vue
│   │   │       ├── RevenueChart.vue
│   │   │       └── DonationRevenueChart.vue
│   │   │
│   │   └── admin/
│   │       ├── AdminUserTable.vue     # User list with hide/restore/delete actions
│   │       ├── AdminUserForm.vue      # Create / Edit user form
│   │       ├── AdminTagTable.vue      # Tag list with CRUD actions
│   │       ├── AdminTagForm.vue       # Create / Edit tag form
│   │       └── AdminStatusBadge.vue   # Badge showing hidden / active / deleted status
│   │
│   ├── composables/            # Reusable Composition API logic
│   │   ├── useAuth.ts          # Login state, token refresh, role check
│   │   ├── useSearch.ts        # Search logic & filters
│   │   ├── usePagination.ts
│   │   ├── useVideoUpload.ts   # Upload progress tracking
│   │   ├── useConfirmDialog.ts
│   │   ├── useTags.ts          # Tag selection logic (guest session + user)
│   │   └── usePaddle.ts        # Paddle.js initialization & checkout open
│   │
│   ├── layouts/                # Layout wrappers
│   │   ├── DefaultLayout.vue   # Header + Sidebar + Main content
│   │   ├── AuthLayout.vue      # Centered card (login/register)
│   │   ├── DashboardLayout.vue # Dashboard sidebar + content area
│   │   └── AdminLayout.vue     # Admin sidebar + content area
│   │
│   ├── router/                 # Vue Router configuration
│   │   ├── index.ts            # Route definitions
│   │   └── guards.ts           # Navigation guards (auth check)
│   │
│   ├── stores/                 # Pinia stores
│   │   ├── authStore.ts        # User auth state, JWT token, role
│   │   ├── videoStore.ts       # Current video, video lists
│   │   ├── channelStore.ts     # Channel data, membership state
│   │   ├── searchStore.ts      # Search query, filters, results
│   │   ├── categoryStore.ts    # Category list
│   │   ├── dashboardStore.ts   # Dashboard analytics data
│   │   ├── tagStore.ts         # Available tags, user selected tags
│   │   ├── donationStore.ts   # Donation state, sent/received lists
│   │   └── adminStore.ts       # Admin user list, tag management
│   │
│   ├── types/                  # TypeScript type definitions
│   │   ├── user.ts             # User, AdminUser (with role, is_hidden)
│   │   ├── video.ts            # Video (with is_hidden, tags)
│   │   ├── channel.ts          # Channel (with is_hidden)
│   │   ├── category.ts
│   │   ├── tag.ts              # Tag, UserTagPreference
│   │   ├── donation.ts         # Donation, CreateDonationPayload (video-level)
│   │   ├── search.ts
│   │   └── api.ts              # API response types
│   │
│   ├── utils/                  # Utility functions
│   │   ├── formatDate.ts
│   │   ├── formatDuration.ts
│   │   ├── formatViews.ts
│   │   └── validators.ts
│   │
│   ├── views/                  # Page-level components (bound to routes)
│   │   ├── LoginView.vue
│   │   ├── HomeView.vue               # Tag-based recommended videos
│   │   ├── SearchResultsView.vue
│   │   ├── CategoryView.vue
│   │   ├── ChannelView.vue
│   │   ├── VideoView.vue
│   │   ├── dashboard/
│   │   │   ├── DashboardView.vue       # Dashboard main wrapper
│   │   │   ├── DashboardVideosView.vue  # My uploaded videos
│   │   │   ├── DashboardUploadView.vue  # Upload new video
│   │   │   ├── DashboardAnalyticsView.vue
│   │   │   ├── DashboardDonationsView.vue # Sent & received donations
│   │   │   └── DashboardSettingsView.vue # Hide/Delete account & channel
│   │   │
│   │   └── admin/
│   │       ├── AdminView.vue           # Admin main wrapper
│   │       ├── AdminUsersView.vue      # User management (list, hide, restore, delete)
│   │       ├── AdminUserEditView.vue   # Create / Edit user
│   │       ├── AdminTagsView.vue       # Tag management (list, create, edit, delete)
│   │       └── AdminTagEditView.vue    # Create / Edit tag
│   │
│   ├── App.vue
│   └── main.ts
│
├── e2e/                        # Playwright E2E tests
│   ├── auth.spec.ts
│   ├── video.spec.ts
│   ├── dashboard.spec.ts
│   ├── tags.spec.ts            # Tag selection & recommendation flow
│   ├── donations.spec.ts      # Donation flow, Paddle checkout, history
│   └── admin.spec.ts           # Admin CRUD, hide/restore/delete
│
├── .env                        # VITE_API_BASE_URL, VITE_PADDLE_*
├── .env.production
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.ts
├── playwright.config.ts
├── vitest.config.ts
└── eslint.config.js
```

---

## Routing

```ts
// router/index.ts
const routes = [
  {
    path: "/login",
    name: "Login",
    component: LoginView,
    meta: { layout: "auth", guest: true },
  },
  {
    path: "/",
    name: "Home",
    component: HomeView,
    meta: { layout: "default" },
  },
  {
    path: "/search",
    name: "SearchResults",
    component: SearchResultsView,
    meta: { layout: "default" },
    // query params: ?q=keyword&category=...&duration=...&date=...&views=...&access=...
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
    path: "/dashboard",
    name: "Dashboard",
    component: DashboardView,
    meta: { layout: "dashboard", requiresAuth: true },
    children: [
      { path: "", name: "DashboardVideos", component: DashboardVideosView },
      {
        path: "upload",
        name: "DashboardUpload",
        component: DashboardUploadView,
      },
      {
        path: "analytics",
        name: "DashboardAnalytics",
        component: DashboardAnalyticsView,
      },
      {
        path: "settings",
        name: "DashboardSettings",
        component: DashboardSettingsView,
      },
      {
        path: "donations",
        name: "DashboardDonations",
        component: DashboardDonationsView,
      },
    ],
  },
  {
    path: "/admin",
    name: "Admin",
    component: AdminView,
    meta: { layout: "admin", requiresAuth: true, requiresAdmin: true },
    children: [
      { path: "", redirect: { name: "AdminUsers" } },
      { path: "users", name: "AdminUsers", component: AdminUsersView },
      {
        path: "users/create",
        name: "AdminUserCreate",
        component: AdminUserEditView,
      },
      {
        path: "users/:id/edit",
        name: "AdminUserEdit",
        component: AdminUserEditView,
      },
      { path: "tags", name: "AdminTags", component: AdminTagsView },
      {
        path: "tags/create",
        name: "AdminTagCreate",
        component: AdminTagEditView,
      },
      {
        path: "tags/:id/edit",
        name: "AdminTagEdit",
        component: AdminTagEditView,
      },
    ],
  },
];
```

### Navigation Guards

```ts
// router/guards.ts
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore();

  // Redirect logged-in users away from login page
  if (to.meta.guest && authStore.isLoggedIn) {
    return next({ name: "Home" });
  }

  // Protect dashboard routes
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    return next({ name: "Login", query: { redirect: to.fullPath } });
  }

  // Protect admin routes — must have role === 'admin'
  if (to.meta.requiresAdmin && authStore.user?.role !== "admin") {
    return next({ name: "Home" });
  }

  // Block hidden users from normal browsing (redirect to contact support)
  if (authStore.isLoggedIn && authStore.user?.is_hidden) {
    if (!to.meta.guest && to.name !== "Login") {
      return next({ name: "Login" });
    }
  }

  next();
});
```

---

## State Management (Pinia Stores)

### authStore

```ts
interface AuthState {
  user: User | null; // includes role: 'user' | 'admin', is_hidden: boolean
  token: string | null;
  isLoggedIn: boolean;
  isAdmin: boolean; // computed: user?.role === 'admin'
}

// Actions: login(), register(), logout(), refreshToken(),
//          hideAccount(), deleteAccount(), deleteChannel()
```

### videoStore

```ts
interface VideoState {
  currentVideo: Video | null;
  recommendedVideos: Video[]; // tag-based recommended
  myVideos: Video[]; // Dashboard uploaded videos
}

// Actions: fetchVideo(), fetchRecommended(tagIDs?), fetchMyVideos(),
//          uploadVideo(), updateVideo(), deleteVideo(), togglePublish()
// Note: fetchRecommended() calls GET /videos/recommended
//       Backend selects random tag combination from user's preferences
```

### tagStore

```ts
interface TagState {
  allTags: Tag[]; // all available tags from server
  selectedTags: Tag[]; // user's picked tags (max 5)
  sessionId: string | null; // for guest users (UUID stored in localStorage)
}

// Actions: fetchAllTags(), fetchMyTags(), setMyTags(tagIDs[]),
//          getOrCreateSessionId()
// Guest flow: if not logged in, generate a UUID sessionId,
//             store in localStorage, pass as query/body param
```

### adminStore

```ts
interface AdminState {
  users: AdminUser[];
  totalUsers: number;
  currentUser: AdminUser | null;
  tags: Tag[]; // for admin tag management
}

interface AdminUser {
  id: number;
  email: string;
  display_name: string;
  role: "user" | "admin";
  is_hidden: boolean;
  created_at: string;
  channel?: { id: number; name: string; is_hidden: boolean };
}

// Actions: fetchUsers(params), getUser(id), createUser(), updateUser(),
//          hideUser(id), restoreUser(id), deleteUser(id),
//          fetchTags(), createTag(), updateTag(), deleteTag()
```

### searchStore

```ts
interface SearchState {
  query: string;
  filters: SearchFilters;
  results: Video[];
  totalCount: number;
  page: number;
}

interface SearchFilters {
  category: number | null;
  durationRange: [number, number] | null; // min/max seconds
  uploadDateRange: [string, string] | null;
  viewCountSort: "asc" | "desc" | null;
  accessType: "public" | "member" | null;
}
```

### channelStore

```ts
interface ChannelState {
  channel: Channel | null;
  isMember: boolean;
  membershipFee: number;
}

// Actions: fetchChannel(), joinMembership(), leaveMembership(), setFee()
```

### dashboardStore

```ts
interface DashboardState {
  analytics: {
    totalViews: { member: number; nonMember: number };
    viewsRanking: VideoRanking[];
    memberCount: number;
    memberRatio: { member: number; nonMember: number };
    revenue: number;
    donationRevenue: number; // total received donations
  };
}
```

### donationStore

```ts
interface DonationState {
  sentDonations: Donation[]; // donations I sent to creators
  receivedDonations: Donation[]; // donations I received on my channel
  loading: boolean;
}

interface Donation {
  id: number;
  video: { id: number; title: string };
  donor: { id: number; display_name: string };
  creator: { id: number; display_name: string; channel_name: string };
  amount: number;
  currency: string;
  message: string;
  paddle_status: "pending" | "completed" | "failed" | "refunded";
  created_at: string;
}

// Actions: createDonation(videoId, amount, currency, message),
//          fetchSentDonations(), fetchReceivedDonations(),
//          openPaddleCheckout(transactionId)  — calls Paddle.js
```

---

## API Layer

### Axios Instance

```ts
// api/index.ts
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL, // e.g. http://localhost:8000
  timeout: 30000,
  headers: { "Content-Type": "application/json" },
});

// Request interceptor — attach JWT
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore().token;
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// Response interceptor — handle 401
apiClient.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      useAuthStore().logout();
      router.push("/login");
    }
    return Promise.reject(err);
  },
);
```

### Endpoint Modules

| Module         | Endpoints                                                                                                                                                                                                                                            |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `auth.ts`      | `POST /auth/login`, `POST /auth/register`, `POST /auth/logout`, `POST /auth/refresh`                                                                                                                                                                 |
| `video.ts`     | `GET /videos/recommended`, `GET /videos/:id`, `POST /videos`, `PUT /videos/:id`, `DELETE /videos/:id`, `PATCH /videos/:id/publish`                                                                                                                   |
| `channel.ts`   | `GET /channels/:id`, `POST /channels/:id/membership`, `DELETE /channels/:id/membership`, `PUT /channels/fee`                                                                                                                                         |
| `category.ts`  | `GET /categories`, `GET /categories/:id/videos`                                                                                                                                                                                                      |
| `dashboard.ts` | `GET /dashboard/videos`, `GET /dashboard/analytics`, `PUT /dashboard/fee`                                                                                                                                                                            |
| `user.ts`      | `PUT /user/display-name`, `PUT /user/password`, `PUT /user/hide`, `DELETE /user/account`, `DELETE /user/channel`                                                                                                                                     |
| `tag.ts`       | `GET /tags`, `GET /tags/my?session_id=...`, `PUT /tags/my`                                                                                                                                                                                           |
| `donation.ts`  | `POST /videos/:id/donate`, `GET /donations/sent`, `GET /donations/received`                                                                                                                                                                                  |
| `admin.ts`     | `GET /admin/users`, `GET /admin/users/:id`, `POST /admin/users`, `PUT /admin/users/:id`, `PUT /admin/users/:id/hide`, `PUT /admin/users/:id/restore`, `DELETE /admin/users/:id`, `POST /admin/tags`, `PUT /admin/tags/:id`, `DELETE /admin/tags/:id` |
| `search`       | `GET /search?q=...&category=...&duration=...&date=...&views=...&access=...`                                                                                                                                                                          |

---

## Key Component Interactions

```
┌─────────────────────────────────────────────────────┐
│                    App.vue                          │
│  ┌───────────────────────────────────────────────┐  │
│  │  Layout (Default / Auth / Dashboard / Admin)  │  │
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

| Page (View)              | Key Components Used                                                                     |
| ------------------------ | --------------------------------------------------------------------------------------- |
| `LoginView`              | `LoginForm`, `RegisterForm`                                                             |
| `HomeView`               | `VideoGrid`, `VideoCard`, `TagSelector`, `SearchBar`                                    |
| `SearchResultsView`      | `VideoGrid`, `VideoCard`, `VideoFilterPanel`, `Pagination`                              |
| `CategoryView`           | `VideoGrid`, `VideoCard`, `Pagination`                                                  |
| `ChannelView`            | `ChannelBanner`, `ChannelVideoList`, `MembershipDialog`                                 |
| `VideoView`              | `VideoPlayer`, `VideoInfo`, `VideoDonateDialog`                                         |
| `DashboardVideosView`    | `DashboardVideoList`, `VideoEditForm`, `ConfirmDialog`                                  |
| `DashboardUploadView`    | `VideoUploadForm`, `TagPicker`                                                          |
| `DashboardAnalyticsView` | `AnalyticsCharts`, `TotalViewsChart`, `ViewsRankingChart`, `DonationRevenueChart`, etc. |
| `DashboardDonationsView` | `DonationHistoryTable`, `Pagination`                                                    |
| `DashboardSettingsView`  | `AccountSettings`, `MembershipFeeForm`, `ConfirmDialog`                                 |
| `AdminUsersView`         | `AdminUserTable`, `AdminStatusBadge`, `ConfirmDialog`, `Pagination`                     |
| `AdminUserEditView`      | `AdminUserForm`                                                                         |
| `AdminTagsView`          | `AdminTagTable`, `ConfirmDialog`, `Pagination`                                          |
| `AdminTagEditView`       | `AdminTagForm`                                                                          |

---

## Auth Flow

```
  User                    Frontend                     Backend
   │                         │                            │
   │── Enter credentials ──▶ │                            │
   │                         │── POST /auth/login ──────▶ │
   │                         │◀── { token, user } ───── │
   │                         │── Save token (Pinia) ──▶   │
   │◀── Redirect to Home ── │                            │
   │                         │                            │
   │── Access Dashboard ──▶  │                            │
   │                         │── Guard checks token ──▶   │
   │                         │── GET /dashboard/** ─────▶ │
   │                         │   (Authorization: Bearer)  │
   │                         │◀── 200 Data ────────────── │
   │◀── Render Dashboard ── │                            │
```

---

## Video Upload Flow

```
  User                    Frontend                     Backend
   │                         │                            │
   │── Fill form + file ───▶ │                            │
   │                         │── POST /videos ──────────▶ │
   │                         │   (multipart/form-data)    │
   │                         │   { file, title, desc,     │
   │                         │     category, isMember }   │
   │                         │                            │
   │◀── Progress bar ─────  │           ┌──────────────┐ │
   │                         │           │ MinIO upload │ │
   │                         │           └──────────────┘ │
   │                         │◀── 201 { video } ────────  │
   │◀── Success toast ─────  │                            │
```

---

## Donation / Paddle Checkout Flow (Video-Level)

Donations are triggered at the **video level** — the "Donate" button appears on the Video Page where the user's intent is strongest (impulse-purchase model).

```
  User (watching video)     Frontend (VideoView)         Backend                  Paddle (Sandbox)
   │                         │                            │                          │
   │── Click "Donate" ─────▶ │                            │                          │
   │── Enter amount+msg ───▶ │                            │                          │
   │                         │── POST /videos/:id/donate ▶ │                          │
   │                         │   { amount, currency,      │                          │
   │                         │     message }              │── CreateTransaction ───▶ │
   │                         │                            │◀── txn_* ID ──────────── │
   │                         │◀── { checkout_url, txn } ─ │                          │
   │                         │                            │                          │
   │                         │── Paddle.Checkout.open({   │                          │
   │                         │     transactionId: txn_*   │                          │
   │                         │   }) ─────────────────────────────────────────▶│
   │◀── Paddle overlay ──── │                            │                          │
   │── Complete payment ──▶  │                            │                          │
   │                         │                            │◀── Webhook: txn.completed │
   │                         │                            │── Update donation status  │
   │◀── Success toast ───── │◀── (poll or ws) ─────────── │                          │
```

---

## Environment Variables

```env
# .env
VITE_API_BASE_URL=http://localhost:8000/api/v1
VITE_APP_TITLE=FenzVideo
VITE_PADDLE_CLIENT_TOKEN=test_...      # Paddle sandbox client-side token
VITE_PADDLE_ENVIRONMENT=sandbox        # sandbox | production

# .env.production
VITE_API_BASE_URL=https://api.fenzvideo.com/api/v1
VITE_PADDLE_CLIENT_TOKEN=live_...      # Paddle production client-side token
VITE_PADDLE_ENVIRONMENT=production
```

---

## Build & Deploy

```bash
# Development
npm run dev          # Vite dev server at localhost:5173

# Production build
npm run build        # Output to dist/
npm run preview      # Preview production build locally

# Testing
npm run test         # Vitest unit tests
npm run test:e2e     # Playwright E2E tests

# Lint & Format
npm run lint         # ESLint
npm run format       # Prettier
```

---

## Key Dependencies (`package.json`)

All dependencies are open source under permissive licenses (MIT / Apache-2.0):

```json
{
  "dependencies": {
    "vue": "^3.5",
    "vue-router": "^4.4",
    "pinia": "^2.2",
    "axios": "^1.7",
    "element-plus": "^2.8",
    "video.js": "^8.17",
    "echarts": "^5.5",
    "vue-echarts": "^7.0",
    "vee-validate": "^4.13",
    "yup": "^1.4",
    "vue-i18n": "^10.0",
    "@iconify/vue": "^4.1",
    "@paddle/paddle-js": "^1.3"
  },
  "devDependencies": {
    "vite": "^6.0",
    "typescript": "^5.6",
    "@vitejs/plugin-vue": "^5.2",
    "tailwindcss": "^3.4",
    "postcss": "^8.4",
    "autoprefixer": "^10.4",
    "vitest": "^2.1",
    "@vue/test-utils": "^2.4",
    "playwright": "^1.48",
    "@playwright/test": "^1.48",
    "eslint": "^9.14",
    "prettier": "^3.4",
    "sass": "^1.80"
  }
}
```

---

## Open-Source Alternatives Reference

| Component       | Current           | Alternative (also open source) |
| --------------- | ----------------- | ------------------------------ |
| UI Framework    | Element Plus      | Naive UI, PrimeVue, Vuetify 3  |
| CSS             | Tailwind CSS      | UnoCSS, Windi CSS              |
| Video Player    | Video.js          | Plyr, Shaka Player, hls.js     |
| Charts          | ECharts           | Chart.js, ApexCharts           |
| Form Validation | VeeValidate + Yup | FormKit, Vuelidate             |
| E2E Testing     | Playwright        | Cypress                        |
| Icons           | Iconify           | Lucide, Heroicons              |
| Build Tool      | Vite              | Rspack, Farm                   |
