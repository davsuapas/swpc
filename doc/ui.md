# WEB USER INTERFACE
The web application is responsive and by default it authenticates against AWS Cognito via OAuth, although it could be authenticated against another provider changing the [configuration](../internal/config/config.go).

Once authenticated, the application shows a dashboard with all the available metrics, as we can see in the figure below.

This dashboard consists of:

- A top bar with a menu of buttons on the right, to perform tasks:
  - Add samples to later create fitted models to predict Chlorine and water quality. Once the model is created, this button can be hidden by changing the [configuration](../internal/config/config.go).
  - Configuration of micro-controller parameters.
  - Refresh chlorine and water quality indicators.
  - Exit application.
- In the main section the screen is divided into two parts:
  - The upper part is made up of several numerical indicators:
    - Water quality: Shows the quality of the water. This indicator is not calculated in real time, it is necessary to push the refresh button in the menu.
    - Chlorine: Shows the chlorine value. This indicator is not calculated in real time, it is necessary to push the refresh button in the menu.
    - Temperature: Shows the temperature value in real time.
    - PH: Shows the PH value in real time.
    - ORP (Oxidation Reduction Potential): Shows the ORP value in real time.
  - The bottom part is made up of several graphs that show the data in real time. These are temperature, PH and ORP.
