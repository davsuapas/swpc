# SWPC - Swimming pool controller

![logo](resources/swpc.jpeg)

The project consists of monitoring the metrics of a swimming pool in real time. The metrics are collected through sensors inside the pool connected to an esp32 board.

The board, depending on the parameters configured in the application via web, sends the metrics information to a server, which broadcasts the metrics to all the connected subscribers.

The metrics obtained by the user are: Temperature, ORP (Oxidation Reduction Potential) and PH. Other metrics such as chlorine and water quality are calculated by two predictive models (artificial intelligence), namely a regression model and a decision tree model.

The metrics can be visualised anytime and anywhere via a web application.

The system is prepared for both microcontroller and server crashes. Therefore, when one of the parts detects that the other is available, the system starts to work normally.

Special care has been taken regarding memory and cpu consumption. Specifically, measurements have been made and for 70 clients connected in real time, values of 20 MB of memory and 0.5% of cpu have been obtained. The system is ready to deploy in aws beanstalk immediately.


[![lint](https://github.com/davsuapas/swpc/workflows/lint/badge.svg)](https://github.com/davsuapas/swpc/actions?query=workflow%3Alint)
[![test](https://github.com/davsuapas/swpc/workflows/test/badge.svg)](https://github.com/davsuapas/swpc/actions?query=workflow%3Atest)
[![codecov](https://codecov.io/github/davsuapas/swpc/branch/main/graph/badge.svg?token=VG71O5HYBA)](https://codecov.io/github/davsuapas/swpc)

![dashboard](./doc/dashboard.gif)

# TABLE OF CONTENTS


1. [Board and sensors](doc/board.md)
2. [Architecture](doc/architecture.md)
3. [Develop](doc/develop.md)
4. [AI](doc/ai.md)
5. [Test](doc/test.md)
6. Deploy

    6.1 [Deploy on AWS](doc/deploy_aws.md)

    6.2 [Deploy on VPS](doc/deploy_vps.md)

7. [Web user interface](doc/ui.md)
8. [ORP Calibration](doc/orpc.md)