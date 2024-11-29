# Use the official Node.js image
FROM node:14

# Set the Current Working Directory inside the container
WORKDIR /nodeapp

# Copy package.json and package-lock.json
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the source code into the container
COPY . .

# Expose port 3000 to the outside world
EXPOSE 3000

# Command to run the application
CMD ["node", "index.js"]

