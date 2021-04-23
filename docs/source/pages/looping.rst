Steps as Loops
##############

Most if not all the times, we want to perform one or more computational steps iteratively based on a list of objects or files.
To this end, Bioflows has abstracted the looping logic away from pipeline authors and it had made it really easy to execute a given
bioflows step or a nested pipeline in a loop. Let's define a simple scenario for a loop.

A Loop Scenario
***************

Assume that we have a directory which contains one or more files, we are interested in looping through that directory
and perform a single computational process on each file individually. Also, assume that there is a tool called "xyztool"
that perform that single computational process on the given file.
Now, because we want to perform a single computational process iteratively on a list of objects/files
a single bioflows tool is sufficient to do the job.

Ok. How can we wrap this logic in Bioflows without writing bash scripts or python scripts to do this....

Please look at the tool definition below.

.. code-block:: yaml

    id: myloop
    name: myloop
    type: tool
    loop: true
    loop_var: "files"
    config:
      - name: files
        type: array
    scripts:
      - type: js
        before: true
        code: >
          self.files = io.ListDir(self.data_dir,true)
    command: "xyztool -f {{files_item}}"

- **First**, we created a step whose type is a "tool" because we will do a single process on each file,To declare a bioflows step as an iterative loop we should add `loop: true` directive and set it to true.

- **Second**, we created a *config* section to hold an internal state variable of type `array` to hold the list of files in the `data_dir`. The parameter shouldn't be in a config section, you can easily make it in *inputs* or *outputs*; It really depends on the author's design decision.

- **Third**, we created a script which is going to execute `before` the execution of the current tool, the script is going to list the contents of the given `data_dir` and return files and/or subdirectories with absolute paths, this is why the second parameter to the function `ListDir` is set to true. the returned list of files is set back to `files` internal variable which is of type `array`

- **Fourth**, we declared `files` variable to be used as the loop variable by setting `loop_var: "files"`.

- **Fifth**, Bioflows creates two implicit dynamic variables,The first one is the variable which will hold the value of the current item in the loop and this variable is prefixed with the name of the `loop_var` which generally has the following naming convention `LOOP-VARIABLE-NAME_item`. The second implicit dynamic variable created is the current loop index and you can access the current index through `loop_index` variable.

Now, assume that we need to perform multiple computational steps against each file individually.
For this, we would need to create a step as a nested pipeline and give it each file accordingly as follows.

First, define the pipeline which will process each file individually, name this file as **xypipeline.yml**

.. code-block:: yaml

    id: files_processing
    name: files_processing
    type: pipeline
    inputs:
      - name: single_file
        type: file
    steps:
      - id: xstep
        name: xstep
        command: "xtool -f {{single_file}}"
      - id: ystep
        name: ystep
        command: "ytool -f {{single_file}}"

Now, create a pipeline which contains the loop

.. code-block:: yaml

    id: myloop
    name: myloop
    type: pipeline
    config:
      - name: files
        type: array
    scripts:
      - type: js
        before: true
        code: >
          self.files = io.ListDir(self.data_dir,true);
    steps:
      - id: process
        name: process
        loop: true
        loop_var: "files"
        url: "file:///xypipeline.yml"
        inputs:
          - id: single_file
            value: "{{files_item}}"












