# ImageResizerX

ImageResizerX is an image resizing and management application developed in Golang. It simplifies the process of resizing images and saving them. The application's design follows the repository pattern for file storage. This means that if you need to change where the resized images are saved (e.g., from the local hard disk to cloud storage or a database), you can do so with relative ease by implementing an alternative repository, currently it is designed to save resized image files on the local hard disk.
This project aims to provide a versatile image processing solution for various applications, all while being efficient and performant.

## Table of Contents
- [Features](#features)
- [Endpoints](#endpoints)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
- [Docker Support](#docker-support)
- [Contributing](#contributing)
- [License](#license)

## Features

ImageResizerX offers the following key features:

- **Image Upload**: Users can upload images via the `/api/v1/upload` endpoint. The application automatically resizes the image and saves it to the database, freeing users from the burden of manual resizing.

- **Image Download**: Resized images can be easily downloaded using the `/api/v1/download/<filename>` endpoint. Users can access their resized images whenever needed.

- **Real-Time Updates**: ImageResizerX uses WebSocket (WS) to broadcast messages when an image has been resized. Clients can connect to the WebSocket server at `/ws/` to receive real-time updates with download links for resized images.

- **User-Friendly-Simple Interface**: The project provides a static home page accessible via `/`. This interface allows users to upload images and connect to the WebSocket for real-time image resizing updates.

## Endpoints

- `/api/v1/upload`: POST endpoint for image upload. It does not wait for the resized image and immediately returns a response.

- `/api/v1/download/<filename>`: GET endpoint to download resized images by providing their unique `image_id`.

- `/ws/`: WebSocket endpoint for real-time updates. It broadcasts messages about the resized images, providing download links.

- `/`: The static home page where users can upload images and connect to the WebSocket for real-time image resizing updates.

## Getting Started

### Prerequisites

Before running ImageResizerX, make sure you have the following prerequisites:

- Golang installed on your system
- Docker and Docker Compose (optional for Docker support)

### Installation

1. Clone this repository to your local machine:

   ```shell
   git clone git@github.com:guiflemes/ImageResizerX.git
   ```

2. Install project dependencies:

   ```shell
   # Navigate to the project directory
   cd imageResizerX

   # Install Golang dependencies
   go get -d ./...
   ```

## Usage

1. Build and run the application:

   ```shell
   # Build the project
   go build

   # Run the application
   ./imageResizerX
   ```

   The application will be available at `http://localhost:8080`.

2. Access the home page (`/`) to upload images and connect to the WebSocket for real-time updates.

3. Use the `/api/v1/upload` endpoint to upload images and the `/api/v1/download/<filename>` endpoint to download resized images by providing their unique `image_id`.

## Docker Support

ImageResizerX can also be run within a Docker container. To do this, make sure you have Docker and Docker Compose installed, and then run:

```shell
docker-compose up
```

The application will be accessible at `http://localhost:8080` just like the local installation.

