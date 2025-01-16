FROM node:23-alpine

WORKDIR /app

# Copy package files
COPY package.json yarn.lock ./

# Install dependencies
RUN yarn install

# Copy source code and config files
COPY . .

# Build the application (this will process Tailwind CSS)
RUN yarn build

# Install serve to run the built app
RUN yarn global add serve

EXPOSE 3000

CMD ["serve", "-s", "build", "-l", "3000"] 