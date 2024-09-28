# Use an official Bun runtime as a parent image
FROM oven/bun:1

# Set the working directory in the container
WORKDIR /app

# Copy package.json and bun.lockb (if you're using bun.lockb)
COPY website/package.json website/bun.lockb* ./

# Install dependencies
RUN bun install --frozen-lockfile

# Copy the rest of the application code
COPY website/ .

# Build the Astro app
RUN bun run build

# Expose the port the app runs on
EXPOSE 4321

# Run the app
CMD ["bun", "run", "start --host"]