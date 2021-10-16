# File system structure

* Root is /app-data
* Each organisation is a separate kubernetes storage class 
<pre>
.
├── organisation_id
    │
    ├── client-files
    │    ├── exam_id
    │        ├── client files
    │
    ├── exam-logs
    │    ├── exam_id
    │        ├── user_id.log
    │
    ├── questions
        ├── question_id
            ├── files
</pre>