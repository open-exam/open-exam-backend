## Connection
1) **client** | message-length (u32) | client-public-key (7 bytes) | auth-jwt (n bytes) |

2) **server** | message-length (u32) | auth-status-code (u32) | server-public-key or error message (7 or n bytes) |

## Connected
* **client** | message-length (u32) | service (u8) | IV (16 bytes) | encrypted-data (21 - message-length bytes) |

* **server** | message-length (u32) | IV (16 bytes) | encrypted-data (20 - message-length bytes) |

## Encrypted data construction
Client starts the counter from 1. A counter of 0 indicates an error message where the request could not be parsed.

* **client** | counter (u32) | request-name (16 bytes) | request-data (n bytes) |

* **server** | counter (u32) | response-data (n bytes) |