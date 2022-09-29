# json2obj

converts a json text file to a data object in the language of your choice

## Usage

To a file

    > json2obj --lang java --output My.java my.json
    file 'my.json' converted into a java class file named My.java

Or to standard out

   > json2obj --lang java my.json
   public class My {
    private long value
    private String id
    public void setId(String id){
        this.id = id;
    }
    public String getId(){
        return this.id
    }
    public void setValue(long value){
        this.value = value;
    }
    public long getValue(){
        return this.value
    }
   }

## How to install

On Mac, Linux or WSL do the following:

    curl -sSfL https://raw.githubusercontent.com/rsvihladremio/ssdownloader/main/script/install | sh 


## How to build

    >./script/build

output goes to bin/json2obj
