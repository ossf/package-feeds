FROM ruby
RUN useradd -ms /bin/bash fakename
USER fakename
RUN mkdir /home/fakename/app
WORKDIR /home/fakename/app
