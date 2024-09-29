# Build stage
FROM oven/bun:1 AS builder
WORKDIR /app
COPY website/package.json website/bun.lockb* ./
RUN bun install --frozen-lockfile
COPY website/ .
RUN bun run build

# Run stage
FROM oven/bun:1-alpine
WORKDIR /app
COPY --from=builder /app/dist /app/dist
COPY --from=builder /app/package.json /app/bun.lockb* ./
RUN bun install --production --frozen-lockfile
EXPOSE 4321
CMD ["bun", "run", "start"]