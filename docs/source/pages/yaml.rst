BioFlows Definition Language
############################

BioFlows Definition Language is based on a declarative and structured markup language called (YAML: Yet Another Markup Language).
You express an individual tool or a complete pipeline/workflow by utilizing predefined set of attributes called "Directives".
These Directives allow you to take full control of tool definition and execution.
In BioFlows, The same directives are used to define both an individual tool or a whole complex pipeline with nested whole pipeline(s).

The following section enumerates all these directives in greater details....

BioFlows Directives
===================
.. csv-table:: BioFlows Tool/Pipeline Directives
   :file:  directives.csv
   :header-rows: 1
   :widths: 130, 170


Maintainer Directive
^^^^^^^^^^^^^^^^^^^^

Maintainer directive describes metadata information about the researcher or the bioinformatician who has written that tool or pipeline.
This person is considered to be in charge and support for this tool or pipeline. Users of the pipeline can use this information to communicate with him.

.. code-block:: yaml

   maintainer:
    username: xxxx
    fullname: xxxxx xxxxx
    email: xxx@xx.com

References Directive
^^^^^^^^^^^^^^^^^^^^

This is an optional list of references. Each reference is an object composed of nested directives.
this directive is used to include references to any scientific publications, papers,
posters and/or articles that might act as additional information sources for users
of this tool and/or pipeline.

You define references directive, as follows....

.. code-block:: yaml

   references:
    - name: "Name of your reference"
      description: "long or short snippet of description about this reference"
      website: http://www.yourreference-url.com
    - name: "Name of your reference"
      description: "long or short snippet of description about this reference"
      website: http://www.yourreference-url.com

Inputs Directive
^^^^^^^^^^^^^^^^

How to define Input Parameters (Inputs)
***************************************

Each separate tool or a tool in a bioinformatics pipeline requires some input(s) parameters
to work with and might or might not produce any output(s). Some Bioflows tools might act as decision steps
or state modifiers in a pipeline and hence these tools will only require some input(s) from previous step(s)
and will not produce any output(s). These tools should be shadowed having ``shadow: true`` in their definition.

In order to define input(s) for a tool or a pipeline, the following is an example inputs definition for a dummy tool..

.. code-block:: yaml

   inputs:
      - type: string
        displayname: The input directory for the command
        description: short or long description of the input file
        name: input_dir
        value: /your/original/dir/location
      - type: string
        displayname: The data directory where the rest of the required files reside
        description: short or long description of the data directory
        name: data_dir
        value: /your/data/dir



The type of the input parameter could be a ``string``, a ``file`` , a ``dir`` or it could be anything else.
It really does not matter the value of this type directive as long as the author of the tool knows how to use it
in either the scripts directive or the command directive. The input `type` directive was added for two reasons; the first
reason would be to act as a fallback directive in case it might be needed in future releases; the second reason would be to help
other readers to better understand the actual type of data that this variable might hold.

Output(s) Directive
^^^^^^^^^^^^^^^^^^^

Output(s) directive defines a set of output parameter(s) which might be produced
by a tool during its execution. the outputs are the actual variables which could be utilized
by other downstream dependent tools in the pipeline. A tool might or might not produce any output(s).
Outputs directive follows the same definition markup as that of the inputs shown above.

.. code-block:: yaml

   outputs:
      - type: file
        displayname: "...."
        description: "...."
        name: output_file
        value: myfile.txt


Notification Directive
^^^^^^^^^^^^^^^^^^^^^^

In complex and long running scientific pipelines, sometimes, we want to be notified about the status of one or more analysis step(s).
The notification in BioFlows happens through sending emails. In order to be notified about a specific task in a pipeline,
you have to add a notification directive within the definition of that particular task specifying three or four attributes
which defines an email [to,cc,title and body] , as follows.

.. code-block:: yaml
    notification:
        to: <The receiver Email Address>
        cc: <an optional directive for a carbon copy to other recipients>
        title: the title of the email
        body: short or long textual description of the email


.. note::
    Please note, for the notification feature to work properly, you have to define proper email settings in BioFlows system configuration section of this documentation.


Capabilities Directive
^^^^^^^^^^^^^^^^^^^^^^

Some Bioinformatics analysis steps require specific computing requirements in terms of how many CPU cores and memory size needed.
For instance, RNA-seq Junction aware aligner ``Hisat2`` requires at least ``160 GB`` of available memory if you need to create
FM index with transcripts from a whole reference genome of an organism taking into account that particular organism SNP recorded
variants. To declare a task with specific computing capabilities, you have to define a capabilities directive within the definition
of the task specifying how many computing cores and memory in Mega Bytes (``MB``) required for the job as follows:

.. code-block:: yaml

    caps:
        cpu: 20 # 20 Cores
        memory: 163840 # 160 GB


By adding a ``caps`` directive in a task, BioFlows master node takes care of executing that particular task onto a suitable computing
cluster node that is able to support both CPU and memory specified.



Scripts Directive
^^^^^^^^^^^^^^^^^

In Scientific computing, especially in Bioinformatics, Pipelines are not fixed chain of steps. These analysis steps have
internal state variables, Input parameters and Output parameters that control the behavior of a given step.
You can control the execution of a given step based on any of its internal state variables using embedded scripting.
In BioFlows, currently, we support a fully compatible ``ECMAScript 6 Javascript Embedded engine`` for writing Javascript code
within a specific pipeline step to control the task internal state. In the future, we will support ``Lua`` as well as ``Python``.

A script in BioFlows is meant to control these internal state variables including Configuration parameters, Input Parameters
as well as Output Parameters. Moreover, when you write a script within a bioflow step, you can control when the script will execute,
either before the current step or after it executes using ``before`` and ``after`` directives.

Example Script:
***************

For a full example usage of a script in a complete pipeline, please check the pipeline example(s) section below.


.. code-block:: yaml

    scripts:
          - type: js
            before: true
            code: >
              var output_file = self.nestedone.remoteTwo.location + "/" + "count.txt";
              var contents = io.ReadFile(output_file);
              self.output_str = "Hello Mohamed, this is the contents of the file : " + contents;

This script is an example embedded JS script within a BioFlows step, It opens a specific generated file in a previous step in the
sample pipeline and it reads the file contents then it writes this contents concatenated with additional text into an output parameter
named ``output_str`` which will be echoed back to the standard output of that particular step.

.. note::
    **io.ReadFile** is not a standard Javascript code library, But instead, we developed a set of custom code libraries in GoLang
    and injected these libraries within the embedded JS virtual machine to make it available for script writers.
    These custom code libraries are developed to perform some lower level OS tasks that Javascript doesn't handle by default.



Furthermore, You can externalize the javascript code into an external `.js` file and refer to this file.

For instance, the above script could be written as follows...

.. code-block:: yaml

    scripts:
          - type: js
            before: true
            file: "file:///concat.js"


Please note that now the javascript code exists in a file named "concat.js", this file exists in the current working directory as
the current main pipeline YAML file.

.. note::
    **file:///** should be followed by a relative file path.


For more information about all the available code libraries, please take a look at Embedded Scripting section of this documentation.



Pipeline Example(s)
===================

Please use the following pipeline as an example to understand how to define the previously explained directives in the table above.

.. code-block:: yaml

    id: secondPipeline
    bioflowId: secondPipeline
    type: pipeline
    name: Second Pipeline
    description:
      -"This tool is the second pipeline"
      -"This tool is the second pipeline"
    website: http://hub.bioflows.io
    version: 1.0.0
    steps:
      - id: 1
        bioflowId: mytool1
        name: Generate
        inputs:
          - type: string
            displayname: The input directory for the command
            name: input_dir
            value: /home/snouto
        outputs:
          - type: file
            name: output_file
            value: myfile.txt
        command: ls -ll {{input_dir}} > {{self_dir}}/{{output_file}}
      - id: 2
        bioflowId: mytool2
        name: Move
        depends: 1
        description: "This is a tool that will list all linux directories"
        website: http://hub.bioflows.io
        inputs:
          - type: file
            displayname: The input file to move
            name: input_file
            value: "{{1.location}}/{{1.output_file}}"
          - type: dir
            name: dest_dir
            description: Destination Directory
            value: "{{self_dir}}/movedFile.txt"
        command: mv {{input_file}} {{dest_dir}}
      - id: 3
        name: count
        depends: 1,2
        command: wc -l {{2.dest_dir}} > {{self_dir}}/count.txt


Another Tool definition....

.. code-block:: yaml

    id: nestedPipeline
    name: nestedPipeline
    type: pipeline
    steps:
      - id: nestedone
        name: nestedone
        url: https://raw.githubusercontent.com/mfawzysami/bioflows/master/scripts/remotepipe.yaml
      - id: nestedtwo
        name: nestedtwo
        depends: nestedone
        command: cp {{second_input_file}} {{self_dir}} && echo "{{output_str}}"
        outputs:
          - type: string
            name: output_str
            description: this file will contain the contents of the count.txt from the previous step
        scripts:
          - type: js
            before: true
            code: >
              var output_file = self.nestedone.remoteTwo.location + "/" + "count.txt";
              var contents = io.ReadFile(output_file);
              self.output_str = "Hello Mohamed, this is the contents of the file : " + contents;

        inputs:
          - type: string
            name: second_input_file
            description: "Second Input File"
            value: "{{nestedone.remoteTwo.location}}/count.txt"
