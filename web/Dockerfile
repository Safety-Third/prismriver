FROM node:16.12.0-alpine3.14 as build

COPY . /srv/
WORKDIR /srv/
RUN yarn install

ARG API_URL
ARG NODE_OPTIONS
ENV VUE_APP_API_URL $API_URL
ENV NODE_OPTIONS $NODE_OPTIONS
RUN yarn run build

FROM nginx:1.21.3-alpine

COPY --from=build /srv/dist/ /usr/share/nginx/html/
