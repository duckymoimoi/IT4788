# ==============================================================
# Hospital Admin Panel - Multi-stage Docker build
# Stage 1: Build React app with Vite
# Stage 2: Serve with Nginx
# ==============================================================

# ---- Build Stage ----
FROM node:20-alpine AS builder
WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci --ignore-scripts

COPY . .

# Build arg for API URL (configurable at build time)
ARG VITE_API_URL=http://group3.it4788.sukkaito.id.vn/api
ENV VITE_API_URL=$VITE_API_URL

RUN npm run build

# ---- Production Stage ----
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
