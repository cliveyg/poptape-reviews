#FROM python:3.7-alpine 
FROM ubuntu:18.04
# as base                                                                                                

RUN mkdir -p /reviews
COPY reviews /reviews
COPY .env /reviews
WORKDIR /reviews

# Install any needed packages specified in requirements.txt

# Make port 8020 available to the world outside this container
EXPOSE 8020

# Run app.py when the container launches
CMD ["./reviews"]
