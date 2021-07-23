FROM buildkite/puppeteer:latest

# Install pulse audio
RUN apt-get -qq update && apt-get install -y pulseaudio

# add root user to group for pulseaudio access
RUN adduser root pulse-access

# Xvfb
RUN apt-get install -y xvfb

# ffmpeg
RUN apt-get install -y ffmpeg

# copy recorder
COPY package.json home/node
COPY src home/node/src
RUN npm install \
    && npm install -g typescript ts-node

COPY entrypoint.sh .
ENTRYPOINT ./entrypoint.sh 