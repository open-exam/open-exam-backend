## Connection
1) **client** | message-length | client-public-key | auth-jwt |

2) **server** | message-length | auth-status-code | server-public-key or error message |

## Connected
* **client** | message-length | service | IV | encrypted-data |

* **server** | message-length | IV | encrypted-data |