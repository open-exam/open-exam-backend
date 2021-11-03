# File system structure

* Root is /app-data
* Each organization_id folder is a separate kubernetes volume
<pre>
.
├── exam-files
│   ├── organization_id
│       ├── exam_id
│             ├── client files
├── exam-logs
│   ├── organization_id
│        ├── exam_id
│             ├── user_id.log
├── questions
│   ├── organization_id
│        ├── question_id
│             ├── files
├── exam-templates
    ├── organization_id
        ├── template_id.json
</pre>