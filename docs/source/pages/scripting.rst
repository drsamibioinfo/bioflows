Embedded Scripting
##################

Introduction
^^^^^^^^^^^^

Bioflows supports embedded scripting. You can access every tool parameters including inputs and outputs and tool command. The idea behind
embedding scripts is to internally change the state of the tool parameters depending on code that will be executed in the context of
each tool either `before` the tool executes or `after` the tool is executed or runs before each iteration within a loop by setting `inloop` directive to true.
There is no limit to how many embedded scripts to include in each tool and this depends on the author of each tool. You can
change the order of their execution by the numeric value you give for `order` directive in each script declaration.

Currently, Bioflows supports Javascript "ECMAScript 6". You can use Javascript to alter the values of the tool parameters according to some conditions.
Besides the power and flexibility that `ECMAScript 6` gives you, there are some other helper functions that could be used to access IO (Input/Output) from within Javascript code.
You will see a lot of examples of how to use scripting to control tool parameters in `Practical Tutorials` section in this documentation.

The following is a simple example of the power of using JS scripting to control the tool parameters..

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


As you can see in the above example, we have read the contents of `count.txt` output file from `remoteTwo` step in `nestedone` task
of the current pipeline through invoking `io.ReadFile` which is a helper function available to Javascript and then we inserted the contents
of that file appended with a custom string to `self.output_str` which is an output parameter used in the command of the current tool.


In the script definition, we have stated that this script should be run before the execution of the current tool command by declaring `before: true` directive (as a pre-command execution script).
On the other hand, you can declare a script to run after the execution of the current tool by declaring `after: true` (as a post-command execution script).

Moreover, you can have multiple pre-command execution scripts and post-command execution scripts. There is no limit to the number
of scripts to include in your tool. You can define the order of pre-command execution scripts or post-command execution scripts through `order:<number>` directive

For instance, you can have two pre-command execution JS scripts, but you need one to be executed before the other in that specific order. you can control
their execution order by declaring `order:1` in the first script and `order: 2` in the second script.....


Helper Functions
^^^^^^^^^^^^^^^^

There are helper functions available to Javascript, these helper functions could be invoked from JS code. The following table contains all the currently supported helper functions.


.. csv-table:: BioFlows Helper Functions
   :file:  funcs.csv
   :header-rows: 1

