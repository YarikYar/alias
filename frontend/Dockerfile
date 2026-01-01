# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Build args
ARG VITE_API_URL
ARG VITE_WS_URL
ARG VITE_BOT_USERNAME

ENV VITE_API_URL=$VITE_API_URL
ENV VITE_WS_URL=$VITE_WS_URL
ENV VITE_BOT_USERNAME=$VITE_BOT_USERNAME

# Copy package files
COPY package.json package-lock.json* ./
RUN npm ci

# Copy source
COPY . .

# Build
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built files
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
