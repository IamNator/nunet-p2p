**Project Documentation**

**Project Overview:**
- **UI**: [Nunet P2P UI](https://iamnator.github.io/nunet-p2p/)
- **Bootstrap Node**: [Nunet Bootstrap Node](https://nunet.verifyotp.io)

**Thought Process:**
- Initially, I considered implementing the solution using a REST API with hard-coded addresses for peer communication. I developed a basic solution based on this approach, which worked effectively.
- However, upon considering scenarios involving more than two machines, I recognized the limitations of this approach. Recalling previous encounters with libp2p, I realized it offered comprehensive solutions for peer discovery and communication.
- With this in mind, I pivoted to building an application utilizing a single pubsub topic. Additionally, I created a simple HTML/JS UI to facilitate interaction with the program. To enhance peer discovery, I incorporated functionality for users to directly connect to specified peers.
- To fully test the solution, I deployed a copy on AWS EC2. Considering Nunet's objective of distributed computing, I recognized the importance of knowing the available computing resources of each node. This led to the consideration of resource-based task assignment.
- For program execution, I implemented a mechanism for the source program to receive a response upon job completion, necessitating the creation of another pubsub topic for responses.

**Other Considerations:**
- I made an effort to minimize reliance on external services such as databases, queues, or pubsub systems like Kafka. This ensures the system's operability in environments with limited or no internet connectivity.

**Limitations and Possible Improvements:**
- Instead of pubsub, employing streams from libp2p could reduce network load.
- Implementing a mechanism to pass the job to another machine if the target machine lacks the necessary resources could improve efficiency.

**Difficulties Encountered:**
- I had limited experience with libp2p initially, but I found the learning and implementation process exciting and rewarding.

---

This documentation provides a comprehensive overview of the project's development process, challenges faced, and potential areas for improvement. It also highlights the iterative nature of the development process and the adaptability required to navigate challenges effectively.


---

**Documentation: Nunet P2P Application**

**Introduction:**
Welcome to the documentation for the Nunet P2P Application. This document aims to provide a comprehensive overview of the project's development process, its architecture, functionality, and future prospects. It is intended for developers, stakeholders, and anyone interested in understanding the inner workings of the Nunet P2P Application.

**Project Overview:**
The Nunet P2P Application is designed to facilitate peer-to-peer communication and distributed computing using libp2p. It allows nodes to discover each other, exchange messages, and collaborate on computing tasks seamlessly. The utilization of libp2p provides robust and efficient peer-to-peer communication, making the application suitable for various decentralized use cases.

**Architecture:**
The architecture of the Nunet P2P Application consists of several key components, including nodes, a bootstrap node, a pubsub system, and a user interface. Nodes communicate with each other using libp2p's protocols, while the bootstrap node serves as an entry point for new nodes to join the network. The pubsub system enables efficient message broadcasting, facilitating communication between nodes. Diagrams illustrating the interactions between these components will be provided for clarity.

**Installation and Setup:**
To set up the Nunet P2P Application, follow these steps:
1. Run the program
2. Connect to the bootstrap node.
3. Configure node settings and parameters.
4. Start the application and join the network.

Detailed instructions for each step will be provided, ensuring a smooth setup process for users.

**User Interface:**
The user interface of the Nunet P2P Application provides users with intuitive controls for interacting with the network. Users can connect to peers, send and receive messages, and monitor network activity. The UI is designed to be user-friendly and accessible, even for users with limited technical knowledge.

**Resource Management:**
The Nunet P2P Application includes functionality for assessing and managing the computing resources of nodes. Tasks are allocated based on resource availability, ensuring efficient utilization of computing resources across the network. This feature enhances the application's scalability and performance, particularly in distributed computing scenarios.

**Communication:**
Communication between nodes in the Nunet P2P Application is facilitated through a pubsub system. Nodes subscribe to topics of interest and publish messages to those topics, allowing for efficient message broadcasting and dissemination. A single pubsub topic is used for general communication, while a separate topic is dedicated to handling responses and acknowledgments.

**Testing:**
The Nunet P2P Application has undergone extensive testing to ensure its reliability, scalability, and performance. Testing methodologies include unit tests, integration tests, and deployment testing on AWS EC2 instances. Various test cases and scenarios were used to validate the application's functionality and robustness.

**Challenges and Learnings:**
Developing the Nunet P2P Application presented several challenges, particularly in working with libp2p. Overcoming these challenges required a deep dive into libp2p's documentation, as well as seeking assistance from online forums and communities. Despite the initial learning curve, the experience proved to be both challenging and rewarding, enhancing my understanding of peer-to-peer networking concepts.

**Limitations and Future Work:**
While the Nunet P2P Application is functional and robust, it has certain limitations that could be addressed in future iterations. For example, employing streams from libp2p could reduce network load, while implementing mechanisms for dynamic task allocation could improve resource utilization further. Future work will focus on addressing these limitations and enhancing the application's capabilities.

**Conclusion:**
In conclusion, the Nunet P2P Application represents a significant milestone in the field of distributed computing. Its use of libp2p enables efficient and scalable peer-to-peer communication, making it suitable for a wide range of decentralized applications. By documenting the development process, challenges faced, and future prospects, this documentation aims to provide valuable insights for developers and stakeholders alike.

**Appendices:**
Additional information, such as code snippets, technical specifications, and diagrams, will be included in the appendices for reference.

- https://www.youtube.com/watch?v=4v-iIB0C9_8&ab_channel=IPFS