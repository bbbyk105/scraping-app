FROM node:20-alpine

WORKDIR /app

COPY package.json package-lock.json* ./
# Use npm install for development (more flexible)
# For production, use npm ci with package-lock.json
RUN npm install --legacy-peer-deps

COPY . .

EXPOSE 3000

CMD ["npm", "run", "dev"]

