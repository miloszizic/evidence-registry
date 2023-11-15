# Digital Evidence Registry



## Overview

The Digital Evidence Registry is a comprehensive solution designed to store and manage evidence for courts across the country. With the increasing reliance on digital evidence in modern litigation, there's a growing need for a unified system to maintain the integrity, authenticity, and accessibility of such evidence.

![SLSA Level 3](https://slsa.dev/images/SLSA-Badge-full-level3.svg)


## Why Use the Digital Evidence Registry?

1. **Centralized Storage**: Eliminate the risk of losing or misplacing crucial evidence by storing all digital evidence in one centralized location.
2. **Integrity & Authenticity**: The registry ensures that all evidence remains untampered with, maintaining its authenticity and ensuring its acceptance in court proceedings.
3. **Easy Retrieval**: With an intuitive search functionality, legal professionals can quickly locate and retrieve the evidence they need.
4. **Accessibility**: Courts across the country can access the system, ensuring consistent and standardized evidence handling.
5. **Security**: State-of-the-art encryption and security protocols are in place to protect sensitive evidence from unauthorized access or breaches.
6. **Reduces Manual Errors**: Digital storage and management significantly reduce the chances of human errors that can compromise the integrity of the evidence.
7. **Cost-Effective**: Over time, a unified digital system can save substantial costs related to evidence handling, storage, and transportation.
8. **Eco-Friendly**: Digital storage is more environmentally friendly than paper-based systems, reducing the carbon footprint of the judicial system.
9. **Scalability**: As the number of cases and volume of evidence grows, the system can scale to accommodate the increasing demand without sacrificing performance or security.


### Installation

```
go get
```
### Test it after installation :
For running the project and testing, you will need to install Docker and Docker Compose.
After you will need to run the following commands to install dependencies and run tests :
```
make tidy
make docker-compose-testing
make documented-tests
```
### Run the project

```
make run
```
