AI Agent Overview
Goal

Build an AI-powered daycare assistant that communicates with prospective families over SMS. Parents can text a dedicated phone number to ask questions about the daycare, receive accurate answers, and eventually schedule a tour.

Knowledge (RAG)

The assistant uses a collection of static knowledge documents (Markdown files) containing information such as programs, hours, tuition, policies, and FAQs. These documents are embedded and retrieved at runtime so the LLM can answer questions accurately and consistently.

Contact Management

All conversations are stored in a database, including contact information, message history, lead status, and important metadata. This provides long-term memory across conversations and allows the assistant to personalize future interactions.

Lead Follow-up

A scheduled job runs daily to identify leads that have not responded or completed the enrollment process. The AI reviews each lead's conversation history, determines whether a follow-up is appropriate, and generates a personalized outreach message to continue moving the family toward scheduling a tour.

Design Principles
Use RAG for factual daycare information.
Use the database as the source of truth for contacts and conversation history.
Keep AI responses helpful, accurate, and conversational.
Automate repetitive follow-up tasks while maintaining a natural, personalized experience.