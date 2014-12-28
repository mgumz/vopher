# vopher - acquire vim-plugins the gopher-way

## usage

![](screenshots/vopher.png)

    $> cd ~/.vim

Acquire all of the plugins referenced in 'vopher.list':

    $> vopher -f vopher.list -dir bundles up
    vopher: (25/25) [=================== 100% ====================]

Let's check for updates / new stuff:

    $> vopher -f vopher.list -dir bundles check
    ## goldenview - https://github.com/zhaocai/GoldenView.Vim#V1.3.5

    - master commits:

      31af855bd5 2014-09-08T08:41:49-04:00 Merge pull request #15 from lllama/patch-1
      0bb152d6d3 2014-08-18T10:55:38+03:00 Update Installation instructions with correct git link
      495a5cef06 2013-10-28T10:18:34-04:00 [admin] update zl
      c6c669b30d 2013-10-28T10:18:34-04:00 [minor] Do not open empty buffer by default
      f726e8885f 2013-10-28T10:18:34-04:00 [minor] tweak minheight
      91d52f86e6 2013-10-28T03:23:52-07:00 Merge pull request #10 from jvanbaarsen/patch-1
      39e97ad002 2013-10-28T10:53:41+01:00 Update README.md
      60f6c3e5a8 2013-05-07T21:42:48-04:00 [admin] update vimup message
      323a1f6c85 2013-05-07T21:17:45-04:00 [admin] update zl.vim
      0b1f325ba0 2013-04-27T20:23:37-04:00 [update] increase GoldenMinHeight
     *c23469a0bc 2013-04-26T17:10:53-04:00 [fix] Dirdiff
      c118d96660 2013-04-26T16:37:33-04:00 [minor] update GoldenViewTrace code
      ...

    - commits:

     *c23469a0bc 2013-04-26T17:10:53-04:00 [fix] Dirdiff
      c118d96660 2013-04-26T16:37:33-04:00 [minor] update GoldenViewTrace code
      ...

    - tags:

     *V1.3.5 2013-04-26T21:11:48Z V1.3.5
      V1.3.0 2013-04-22T21:57:01Z V1.3.0
      ...

'GoldenView.Vim#V1.3.5' is referenced in the vopher.list-file. vopher tries to
guess what commit this actually is and marks that line with a '\*'. so, you can
easily see that there seems to be no new release for 'GoldenView', allthough
there are some new commits.


So, what's in my 'bundles' directory and how do they relate to my
vopher.list-file?

    $> vopher -f vopher.list -dir bundles status
    v  EasyDigraph.vim
       buftabs
    v  calendar-vim
    v  csv.vim
       ctrlp.vim
       emmet-vim
    v  goldenview
    v  gundo.vim
       lightline.vim
       pyflakes-vim
    v  scrollfix
       syntastic
    v  tagbar
    v  taglist-46
    v  unicode.vim
       vim-bbye
       vim-bufferline
       vim-colors-solarized
    vm vim-fugitive
       vim-gitgutter
    v  vim-go
       vim-jinja
    v  vim-largefile
    v  vim-markdown
    v  vim-ps1

Lines marked with 'v' are plugins referenced in the vopher.list. 'vm' marked
lines are referenced plugins which are missing (acquire them by using the
'update' action). lines without a special prefix are folders inside 'bundles'
but they are not handled by vopher.

I need more color! Are there any colorschemes available?

    $> vopher search colors
    5856 vim-colors-solarized precision colorscheme for the vim text editor
       github: https://github.com/altercation/vim-colors-solarized

    1151 vim-colorschemes one colorscheme pack to rule them all!
       github: https://github.com/flazz/vim-colorschemes

    482 vividchalk.vim vividchalk.vim: a colorscheme strangely reminiscent of Vibrant Ink for a certain OS X editor
          vim: http://www.vim.org/scripts/script.php?script_id=1891
       github: https://github.com/tpope/vim-vividchalk

    459 vim-css-color Highlight colors in css files
       github: https://github.com/ap/vim-css-color

    416 unite-colorscheme A unite.vim plugin
          vim: http://www.vim.org/scripts/script.php?script_id=3318
       github: https://github.com/ujihisa/unite-colorscheme
    ...

## actions

TODO

## the vopher-file format

the vopher-file is pretty simple:

    # a comment starts with a '#'
    # empty lines are ignored

    # fetches vim-fugitive, current HEAD
    https://github.com/tpope/vim-fugitive

    # fetches vim-fugitive, tagged release 'v2.1'
    https://github.com/tpope/vim-fugitive#v2.1.zip

    # fetches vim-fugitive, name the output-folder 'foo'
    foo https://github.com/tpope/vim-fugitive

    # fetches vim-fugitive, apply some options to the
    # parser/extractor
    https://github.com/tpope/vim-fugitive option1=foo option2=bar

## faq

> why??

pathogen (which is what i use) has no means on it's own to acquire plugins.

vundle needs git. it fetches the whole history of any plugin. i am not
interested in the history, i am just interested in a certain snapshot for
a certain vim-plugin. in addition to that: the git installation on windows
takes up ~ 250mb. all of my vim-plugin take up ~ 4mb.

neobundle needs git or svn.

> but why not use curl?? or python??? or ruby???

curl is easy to install and available everywhere. but it's a bit stupid on
it's own. i would have to write a lot of what `vopher` does on it's own in a
real programming language 'x'. or vimscript. which would lead to even more
code and maybe an additional interpreter which might need even more stuff. on
windows the curl-binary which supports https weighs ~ 1.6mb. a python
installer for windows weighs ~ 17mb, installed ~ 60mb. yeah, one could create
a standalone binary with something like pyinstaller and then we are not better
off than just doing it via golang and it's builtin network- and concurrent
powers.

> but python and ruby are just a `brew install` away?

yep. if you are working mostly on the same platform you can get very
comfortable with your nice and cosy environment. if you switch platform
borders on a regular basis things become a bit more complicated. i want to
place one .zip file on my server, containing all my vim-files, vopher-binaries
and then i am ready to go in no time.

> there is no vim-integration!!

yep, nothing to see on this front here .. yet. i need some means to exchange
messages between vim and vopher in an asyncronous way. without the need for
+clientserver. maybe via a named pipe.

## license

Copyright (c) Mathias Gumz. Distributed under the same terms as Vim itself.
See :help license.
