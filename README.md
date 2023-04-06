# one-step-project-server
## Overview
This is the server side code of an application that hits the one step gps devices api, fetches the response and extracts
meaningful information from the api and provides multiple apis and functionality to interact with the data. The 
following are the list of APIs supported by the server side of the app.
1. GET /devices?page= - This is a get request that returns the list of devices with info like name, device id, active state, online status, drive status, latitude, longitude and altitude. The responses are sorted based on user preferences, and API also accepts a page argument which returns paginated responses.
2. POST /preferences - This is an API to update the user preferences and individual device preferences. User preferences include sort column, sort order and number of rows for pagination. Individual device preferences include icon for the device and option to hide the device from the devices api response.
3. GET /preferences - This is an API to retrieves the stored preferences and returns it back in the response. The preferences are same as above.
4. POST /upload?device_id= - This is an API used to upload an image to the server. This is the icon which will get associated with the device_id.
5. GET /image/:image_path - This is an API that returns the image in the path provided.

## How to run the program
1. Clone this repository.
2. Set the *API_KEY* environment variable with the corresponding value for the one step api key.
3. From the root folder, run the command *go build*, this will generate an executable file.
4. Run the executable file to start the server
